use reqwest::Client;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct ApiResponse<T> {
    pub success: bool,
    pub data: Option<T>,
    pub error: Option<String>,
}

#[derive(Debug, Clone)]
pub struct ApiClient {
    base_url: String,
    client: Client,
}

impl ApiClient {
    pub fn new(base_url: &str) -> Self {
        Self {
            base_url: base_url.to_string(),
            client: Client::new(),
        }
    }

    fn url(&self, path: &str) -> String {
        format!("{}{}", self.base_url, path)
    }

    pub async fn get<T: for<'de> Deserialize<'de>>(&self, path: &str) -> Result<T, String> {
        let resp = self
            .client
            .get(self.url(path))
            .send()
            .await
            .map_err(|e| format!("请求失败: {}", e))?;

        let api_resp: ApiResponse<T> = resp
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        if api_resp.success {
            api_resp.data.ok_or_else(|| "响应数据为空".to_string())
        } else {
            Err(api_resp.error.unwrap_or_else(|| "未知错误".to_string()))
        }
    }

    pub async fn post<T: for<'de> Deserialize<'de>, B: Serialize>(
        &self,
        path: &str,
        body: &B,
    ) -> Result<T, String> {
        let resp = self
            .client
            .post(self.url(path))
            .json(body)
            .send()
            .await
            .map_err(|e| format!("请求失败: {}", e))?;

        let api_resp: ApiResponse<T> = resp
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        if api_resp.success {
            api_resp.data.ok_or_else(|| "响应数据为空".to_string())
        } else {
            Err(api_resp.error.unwrap_or_else(|| "未知错误".to_string()))
        }
    }

    pub async fn put<T: for<'de> Deserialize<'de>, B: Serialize>(
        &self,
        path: &str,
        body: &B,
    ) -> Result<T, String> {
        let resp = self
            .client
            .put(self.url(path))
            .json(body)
            .send()
            .await
            .map_err(|e| format!("请求失败: {}", e))?;

        let api_resp: ApiResponse<T> = resp
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        if api_resp.success {
            api_resp.data.ok_or_else(|| "响应数据为空".to_string())
        } else {
            Err(api_resp.error.unwrap_or_else(|| "未知错误".to_string()))
        }
    }

    pub async fn put_empty<T: for<'de> Deserialize<'de>>(&self, path: &str) -> Result<T, String> {
        let resp = self
            .client
            .put(self.url(path))
            .send()
            .await
            .map_err(|e| format!("请求失败: {}", e))?;

        let api_resp: ApiResponse<T> = resp
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        if api_resp.success {
            api_resp.data.ok_or_else(|| "响应数据为空".to_string())
        } else {
            Err(api_resp.error.unwrap_or_else(|| "未知错误".to_string()))
        }
    }

    pub async fn delete<T: for<'de> Deserialize<'de>>(&self, path: &str) -> Result<T, String> {
        let resp = self
            .client
            .delete(self.url(path))
            .send()
            .await
            .map_err(|e| format!("请求失败: {}", e))?;

        let api_resp: ApiResponse<T> = resp
            .json()
            .await
            .map_err(|e| format!("解析响应失败: {}", e))?;

        if api_resp.success {
            api_resp.data.ok_or_else(|| "响应数据为空".to_string())
        } else {
            Err(api_resp.error.unwrap_or_else(|| "未知错误".to_string()))
        }
    }
}
