pub mod error;
pub mod event;
pub mod lexer;
pub mod parser;

pub use error::{Error, Result};
pub use event::{Event, Number, Position};
pub use parser::{parse, parse_with_config, Config};

use std::collections::HashSet;
use std::fs::File;
use std::io::BufReader;

pub fn json_keys<P: AsRef<std::path::Path>>(path: P) -> Result<Vec<String>> {
    let file = File::open(path)?;
    let reader = BufReader::new(file);
    let mut keys = Vec::new();
    let mut seen = HashSet::new();

    parse(reader, |event| {
        if let Event::Key(key, _) = event {
            let key_str = key.to_string();
            if !seen.contains(&key_str) {
                seen.insert(key_str.clone());
                keys.push(key_str);
            }
        }
    })?;

    Ok(keys)
}

pub fn json_sum_numbers<P: AsRef<std::path::Path>>(path: P) -> Result<f64> {
    let file = File::open(path)?;
    let reader = BufReader::new(file);
    let mut sum = 0.0;

    parse(reader, |event| {
        if let Event::Number(num, _) = event {
            sum += num.to_f64();
        }
    })?;

    Ok(sum)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::event::Event;
    use crate::event::Number::*;

    #[derive(Debug, Clone, PartialEq)]
    enum OwnedEvent {
        ObjectStart(Position),
        ObjectEnd(Position),
        ArrayStart(Position),
        ArrayEnd(Position),
        Key(String, Position),
        String(String, Position),
        Number(crate::event::Number, Position),
        Bool(bool, Position),
        Null(Position),
    }

    fn collect_events(json: &str) -> Vec<OwnedEvent> {
        let mut events = Vec::new();
        parse(json.as_bytes(), |event| {
            events.push(match event {
                Event::ObjectStart(p) => OwnedEvent::ObjectStart(p),
                Event::ObjectEnd(p) => OwnedEvent::ObjectEnd(p),
                Event::ArrayStart(p) => OwnedEvent::ArrayStart(p),
                Event::ArrayEnd(p) => OwnedEvent::ArrayEnd(p),
                Event::Key(k, p) => OwnedEvent::Key(k.to_string(), p),
                Event::String(s, p) => OwnedEvent::String(s.to_string(), p),
                Event::Number(n, p) => OwnedEvent::Number(n, p),
                Event::Bool(b, p) => OwnedEvent::Bool(b, p),
                Event::Null(p) => OwnedEvent::Null(p),
            });
        }).unwrap();
        events
    }

    fn parse_must_fail(json: &str) -> bool {
        let result = parse(json.as_bytes(), |_| {});
        result.is_err()
    }

    #[test]
    fn test_empty_object() {
        let events = collect_events("{}");
        assert!(matches!(events[0], OwnedEvent::ObjectStart(_)));
        assert!(matches!(events[1], OwnedEvent::ObjectEnd(_)));
        assert_eq!(events.len(), 2);
    }

    #[test]
    fn test_empty_array() {
        let events = collect_events("[]");
        assert!(matches!(events[0], OwnedEvent::ArrayStart(_)));
        assert!(matches!(events[1], OwnedEvent::ArrayEnd(_)));
        assert_eq!(events.len(), 2);
    }

    #[test]
    fn test_simple_object() {
        let events = collect_events(r#"{"name": "test", "value": 42}"#);
        assert!(matches!(events[0], OwnedEvent::ObjectStart(_)));
        assert!(matches!(&events[1], OwnedEvent::Key(k, _) if k == "name"));
        assert!(matches!(&events[2], OwnedEvent::String(s, _) if s == "test"));
        assert!(matches!(&events[3], OwnedEvent::Key(k, _) if k == "value"));
        assert!(matches!(&events[4], OwnedEvent::Number(Integer(42), _)));
        assert!(matches!(events[5], OwnedEvent::ObjectEnd(_)));
    }

    #[test]
    fn test_number_types() {
        let events = collect_events(r#"[42, 3.14, -10, 1e5]"#);
        
        match &events[1] {
            OwnedEvent::Number(Integer(42), _) => {}
            e => panic!("Expected Integer(42), got {:?}", e),
        }
        
        match &events[2] {
            OwnedEvent::Number(Float(f), _) if (*f - 3.14).abs() < 0.0001 => {}
            e => panic!("Expected Float(3.14), got {:?}", e),
        }
        
        match &events[3] {
            OwnedEvent::Number(Integer(-10), _) => {}
            e => panic!("Expected Integer(-10), got {:?}", e),
        }
        
        match &events[4] {
            OwnedEvent::Number(Float(f), _) if (*f - 100000.0).abs() < 0.0001 => {}
            e => panic!("Expected Float(100000), got {:?}", e),
        }
    }

    #[test]
    fn test_string_escapes() {
        let events = collect_events(r#"["hello\nworld", "tab\there", "quote\"test", "backslash\\path"]"#);
        
        match &events[1] {
            OwnedEvent::String(s, _) if s == "hello\nworld" => {}
            e => panic!("Expected 'hello\\nworld', got {:?}", e),
        }
        
        match &events[2] {
            OwnedEvent::String(s, _) if s == "tab\there" => {}
            e => panic!("Expected 'tab\\there', got {:?}", e),
        }
        
        match &events[3] {
            OwnedEvent::String(s, _) if s == "quote\"test" => {}
            e => panic!("Expected 'quote\"test', got {:?}", e),
        }
        
        match &events[4] {
            OwnedEvent::String(s, _) if s == "backslash\\path" => {}
            e => panic!("Expected 'backslash\\\\path', got {:?}", e),
        }
    }

    #[test]
    fn test_unicode_escape() {
        let events = collect_events(r#"["\u4e2d\u6587"]"#);
        
        match &events[1] {
            OwnedEvent::String(s, _) if s == "中文" => {}
            e => panic!("Expected '中文', got {:?}", e),
        }
    }

    #[test]
    fn test_boolean_and_null() {
        let events = collect_events(r#"[true, false, null]"#);
        
        assert!(matches!(&events[1], OwnedEvent::Bool(true, _)));
        assert!(matches!(&events[2], OwnedEvent::Bool(false, _)));
        assert!(matches!(&events[3], OwnedEvent::Null(_)));
    }

    #[test]
    fn test_trailing_comma_array_allowed() {
        let events = collect_events(r#"[1, 2, 3,]"#);
        assert!(matches!(&events[1], OwnedEvent::Number(Integer(1), _)));
        assert!(matches!(&events[2], OwnedEvent::Number(Integer(2), _)));
        assert!(matches!(&events[3], OwnedEvent::Number(Integer(3), _)));
    }

    #[test]
    fn test_trailing_comma_object_allowed() {
        let events = collect_events(r#"{"a": 1, "b": 2,}"#);
        assert!(matches!(&events[1], OwnedEvent::Key(k, _) if k == "a"));
        assert!(matches!(&events[2], OwnedEvent::Number(Integer(1), _)));
        assert!(matches!(&events[3], OwnedEvent::Key(k, _) if k == "b"));
        assert!(matches!(&events[4], OwnedEvent::Number(Integer(2), _)));
    }

    #[test]
    fn test_leading_comma_array_rejected() {
        assert!(parse_must_fail(r#"[,1,2]"#));
    }

    #[test]
    fn test_leading_comma_object_rejected() {
        assert!(parse_must_fail(r#"{,"a":1}"#));
    }

    #[test]
    fn test_double_comma_array_rejected() {
        assert!(parse_must_fail(r#"[1,,2]"#));
    }

    #[test]
    fn test_double_comma_object_rejected() {
        assert!(parse_must_fail(r#"{"a":1,,"b":2}"#));
    }

    #[test]
    fn test_nested_depth_under_limit() {
        let mut json = String::new();
        for _ in 0..10 {
            json.push_str("[");
        }
        for _ in 0..10 {
            json.push_str("]");
        }
        let events = collect_events(&json);
        assert_eq!(events.len(), 20);
    }

    #[test]
    fn test_nested_depth_exceeds_limit() {
        let mut json = String::new();
        for _ in 0..130 {
            json.push_str("[");
        }
        for _ in 0..130 {
            json.push_str("]");
        }
        assert!(parse_must_fail(&json));
    }

    #[test]
    fn test_large_integer_preserves_precision() {
        let max_i64 = i64::MAX.to_string();
        let json = format!("[{}]", max_i64);
        let events = collect_events(&json);
        
        match &events[1] {
            OwnedEvent::Number(Integer(n), _) if *n == i64::MAX => {}
            e => panic!("Expected Integer(i64::MAX), got {:?}", e),
        }
    }

    #[test]
    fn test_position_tracking() {
        let json = r#"{
  "name": "test"
}"#;
        let events = collect_events(json);
        
        match &events[0] {
            OwnedEvent::ObjectStart(p) => assert_eq!((p.line, p.column), (1, 1)),
            _ => panic!("Expected ObjectStart"),
        }
        
        match &events[1] {
            OwnedEvent::Key(_, p) => assert_eq!((p.line, p.column), (2, 3)),
            _ => panic!("Expected Key"),
        }
    }

    #[test]
    fn test_strict_mode_rejects_trailing_comma() {
        let json = r#"[1, 2,]"#;
        let result = parse_with_config(
            json.as_bytes(),
            |_| {},
            Config {
                max_depth: 128,
                allow_trailing_comma: false,
            },
        );
        assert!(result.is_err());
    }
}
