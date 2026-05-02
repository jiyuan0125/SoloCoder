pub fn matches(topic_filter: &str, topic_name: &str) -> bool {
    let filter_parts: Vec<&str> = topic_filter.split('/').collect();
    let topic_parts: Vec<&str> = topic_name.split('/').collect();

    let mut filter_iter = filter_parts.iter();
    let mut topic_iter = topic_parts.iter();

    loop {
        match (filter_iter.next(), topic_iter.next()) {
            (Some(&"#"), _) => {
                return true;
            }
            (Some(&"+"), Some(_)) => {
                continue;
            }
            (Some(filter_part), Some(topic_part)) => {
                if filter_part != topic_part {
                    return false;
                }
            }
            (None, None) => {
                return true;
            }
            _ => {
                return false;
            }
        }
    }
}

pub fn validate_topic_filter(filter: &str) -> bool {
    if filter.is_empty() {
        return false;
    }
    
    let parts: Vec<&str> = filter.split('/').collect();
    
    for (i, part) in parts.iter().enumerate() {
        if part.contains('#') {
            if *part != "#" {
                return false;
            }
            if i != parts.len() - 1 {
                return false;
            }
        }
        if part.contains('+') && *part != "+" {
            return false;
        }
    }
    
    true
}

pub fn validate_topic_name(name: &str) -> bool {
    if name.is_empty() {
        return false;
    }
    if name.contains('+') || name.contains('#') {
        return false;
    }
    true
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_exact_match() {
        assert!(matches("a/b/c", "a/b/c"));
        assert!(!matches("a/b/c", "a/b/d"));
        assert!(!matches("a/b", "a/b/c"));
        assert!(!matches("a/b/c", "a/b"));
    }

    #[test]
    fn test_plus_wildcard() {
        assert!(matches("sensor/+/temperature", "sensor/kitchen/temperature"));
        assert!(matches("sensor/+/temperature", "sensor/livingroom/temperature"));
        assert!(!matches("sensor/+/temperature", "sensor/kitchen/humidity"));
        assert!(!matches("sensor/+/temperature", "sensor/kitchen/room/temperature"));
        assert!(matches("+/+/+", "a/b/c"));
        assert!(matches("a/+", "a/b"));
        assert!(!matches("a/+", "a/b/c"));
    }

    #[test]
    fn test_hash_wildcard() {
        assert!(matches("sensor/#", "sensor/kitchen/temperature"));
        assert!(matches("sensor/#", "sensor/kitchen/temperature/humidity"));
        assert!(matches("sensor/#", "sensor"));
        assert!(!matches("sensor/#", "sensors/kitchen"));
        assert!(matches("#", "any/topic/name"));
        assert!(matches("#", "single"));
        assert!(matches("sport/tennis/#", "sport/tennis/player1"));
        assert!(matches("sport/tennis/#", "sport/tennis"));
    }

    #[test]
    fn test_combined_wildcards() {
        assert!(matches("sport/+/player1/#", "sport/tennis/player1/score"));
        assert!(matches("sport/+/player1/#", "sport/football/player1"));
        assert!(!matches("sport/+/player1/#", "sport/tennis/player2"));
    }

    #[test]
    fn test_validate_filter() {
        assert!(validate_topic_filter("a/b/c"));
        assert!(validate_topic_filter("a/+/c"));
        assert!(validate_topic_filter("a/#"));
        assert!(validate_topic_filter("+/b/#"));
        assert!(validate_topic_filter("#"));
        assert!(!validate_topic_filter("a/#/b"));
        assert!(!validate_topic_filter("a/#+"));
        assert!(!validate_topic_filter(""));
    }

    #[test]
    fn test_validate_topic_name() {
        assert!(validate_topic_name("a/b/c"));
        assert!(validate_topic_name("sensor/kitchen/temperature"));
        assert!(!validate_topic_name("sensor/+/temperature"));
        assert!(!validate_topic_name("sensor/#"));
        assert!(!validate_topic_name(""));
    }
}
