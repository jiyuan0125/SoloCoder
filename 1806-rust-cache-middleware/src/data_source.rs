use dashmap::DashMap;
use reqwest::StatusCode;
use serde_json::Value;
use std::collections::HashMap;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum DataSourceError {
    #[error("data source already exists")]
    AlreadyExists,
    #[error("data source not found")]
    NotFound,
    #[error("http error: {0}")]
    Http(#[from] reqwest::Error),
}

pub struct DataSource {
    pub name: String,
    pub url_template: String,
    pub method: String,
}

pub struct DataSourceManager {
    sources: DashMap<String, DataSource>,
    client: reqwest::Client,
}

impl DataSourceManager {
    pub fn new() -> Self {
        Self {
            sources: DashMap::new(),
            client: reqwest::Client::new(),
        }
    }

    pub fn create(
        &self,
        name: String,
        url_template: String,
        method: String,
    ) -> Result<(), DataSourceError> {
        if self.sources.contains_key(&name) {
            return Err(DataSourceError::AlreadyExists);
        }
        self.sources.insert(
            name.clone(),
            DataSource {
                name,
                url_template,
                method,
            },
        );
        Ok(())
    }

    pub fn get(&self, name: &str) -> Option<DataSource> {
        self.sources.get(name).map(|entry| DataSource {
            name: entry.name.clone(),
            url_template: entry.url_template.clone(),
            method: entry.method.clone(),
        })
    }

    pub fn list(&self) -> Vec<DataSource> {
        self.sources
            .iter()
            .map(|entry| DataSource {
                name: entry.name.clone(),
                url_template: entry.url_template.clone(),
                method: entry.method.clone(),
            })
            .collect()
    }

    pub fn update(
        &self,
        name: &str,
        url_template: String,
        method: Option<String>,
    ) -> Result<(), DataSourceError> {
        let mut entry = self
            .sources
            .get_mut(name)
            .ok_or(DataSourceError::NotFound)?;
        entry.url_template = url_template;
        if let Some(m) = method {
            entry.method = m;
        }
        Ok(())
    }

    pub fn delete(&self, name: &str) -> Result<(), DataSourceError> {
        self.sources
            .remove(name)
            .map(|_| ())
            .ok_or(DataSourceError::NotFound)
    }

    fn render_url(template: &str, params: &HashMap<String, String>) -> String {
        let mut result = template.to_string();
        for (key, value) in params {
            result = result.replace(&format!("{{{}}}", key), value);
        }
        result
    }

    pub async fn fetch(
        &self,
        source: &DataSource,
        params: &HashMap<String, String>,
    ) -> Result<(Value, StatusCode), DataSourceError> {
        let url = Self::render_url(&source.url_template, params);

        let method = match source.method.to_uppercase().as_str() {
            "POST" => reqwest::Method::POST,
            "PUT" => reqwest::Method::PUT,
            "DELETE" => reqwest::Method::DELETE,
            _ => reqwest::Method::GET,
        };

        let response = self.client.request(method, &url).send().await?;
        let status = response.status();

        let data: Value = response.json().await.unwrap_or_else(|_| Value::Null);

        Ok((data, status))
    }
}
