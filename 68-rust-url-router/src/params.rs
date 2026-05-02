use std::collections::HashMap;

#[derive(Debug, Clone, Default)]
pub struct Params {
    inner: HashMap<String, String>,
}

impl Params {
    pub fn new() -> Self {
        Params {
            inner: HashMap::new(),
        }
    }

    pub fn insert(&mut self, key: impl Into<String>, value: impl Into<String>) {
        self.inner.insert(key.into(), value.into());
    }

    pub fn get(&self, key: &str) -> Option<&String> {
        self.inner.get(key)
    }

    pub fn get_str(&self, key: &str) -> Option<&str> {
        self.inner.get(key).map(|s| s.as_str())
    }

    pub fn len(&self) -> usize {
        self.inner.len()
    }

    pub fn is_empty(&self) -> bool {
        self.inner.is_empty()
    }

    pub fn iter(&self) -> std::collections::hash_map::Iter<'_, String, String> {
        self.inner.iter()
    }
}

impl IntoIterator for Params {
    type Item = (String, String);
    type IntoIter = std::collections::hash_map::IntoIter<String, String>;

    fn into_iter(self) -> Self::IntoIter {
        self.inner.into_iter()
    }
}

impl<'a> IntoIterator for &'a Params {
    type Item = (&'a String, &'a String);
    type IntoIter = std::collections::hash_map::Iter<'a, String, String>;

    fn into_iter(self) -> Self::IntoIter {
        self.inner.iter()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_params_insert_and_get() {
        let mut params = Params::new();
        params.insert("id", "42");
        params.insert("post_id", "10");

        assert_eq!(params.get("id"), Some(&"42".to_string()));
        assert_eq!(params.get("post_id"), Some(&"10".to_string()));
        assert_eq!(params.get("nonexistent"), None);
    }

    #[test]
    fn test_params_len() {
        let mut params = Params::new();
        assert_eq!(params.len(), 0);
        assert!(params.is_empty());

        params.insert("id", "1");
        assert_eq!(params.len(), 1);
        assert!(!params.is_empty());

        params.insert("name", "test");
        assert_eq!(params.len(), 2);
    }

    #[test]
    fn test_params_iter() {
        let mut params = Params::new();
        params.insert("id", "1");
        params.insert("name", "test");

        let keys: Vec<_> = params.iter().map(|(k, _)| k.clone()).collect();
        assert!(keys.contains(&"id".to_string()));
        assert!(keys.contains(&"name".to_string()));
    }
}
