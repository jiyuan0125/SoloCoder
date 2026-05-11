use anyhow::{Context, Result};
use chrono::{DateTime, Utc};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Serialize, Deserialize)]
pub struct HealthResponse {
    pub status: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ProductResponse {
    pub id: Uuid,
    pub name: String,
    pub original_price: u64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ActivitySummaryResponse {
    pub id: Uuid,
    pub product_name: String,
    pub original_price: u64,
    pub deposit_amount: u64,
    pub inflated_amount: u64,
    pub final_amount: u64,
    pub inflation_rate: u32,
    pub max_participants: u32,
    pub current_participants: u32,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub status: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct OrderDetailResponse {
    pub id: Uuid,
    pub user_id: String,
    pub activity_id: Uuid,
    pub product_name: String,
    pub original_price: u64,
    pub deposit_amount: u64,
    pub inflated_amount: u64,
    pub final_amount: u64,
    pub status: String,
    pub deposit_paid_time: Option<DateTime<Utc>>,
    pub final_paid_time: Option<DateTime<Utc>>,
    pub final_payment_deadline: Option<DateTime<Utc>>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct SimpleResponse {
    pub success: bool,
}

#[derive(Debug, Serialize)]
struct CreateProductRequest {
    name: String,
    original_price: u64,
}

#[derive(Debug, Serialize)]
struct CreateActivityRequest {
    product_id: Uuid,
    deposit_amount: u64,
    inflation_rate: u32,
    max_participants: u32,
    start_time: DateTime<Utc>,
    end_time: DateTime<Utc>,
    #[serde(skip_serializing_if = "Option::is_none")]
    final_payment_deadline_hours: Option<i64>,
}

#[derive(Debug, Serialize)]
struct PayDepositRequest {
    user_id: String,
    activity_id: Uuid,
}

#[derive(Debug, Serialize)]
struct PayFinalRequest {
    paid_amount: u64,
}

pub struct PresaleApiClient {
    base_url: String,
    client: Client,
}

impl PresaleApiClient {
    pub fn new(base_url: &str) -> Self {
        Self {
            base_url: base_url.trim_end_matches('/').to_string(),
            client: Client::new(),
        }
    }

    fn url(&self, path: &str) -> String {
        format!("{}{}", self.base_url, path)
    }

    pub async fn health(&self) -> Result<HealthResponse> {
        let resp = self
            .client
            .get(self.url("/health"))
            .send()
            .await
            .context("Failed to send health check request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("Health check failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<HealthResponse>()
            .await
            .context("Failed to parse health response")?;

        Ok(result)
    }

    pub async fn create_product(&self, name: &str, price: u64) -> Result<ProductResponse> {
        let req = CreateProductRequest {
            name: name.to_string(),
            original_price: price,
        };

        let resp = self
            .client
            .post(self.url("/products"))
            .json(&req)
            .send()
            .await
            .context("Failed to send create product request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("Create product failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<ProductResponse>()
            .await
            .context("Failed to parse product response")?;

        Ok(result)
    }

    pub async fn list_products(&self) -> Result<Vec<ProductResponse>> {
        let resp = self
            .client
            .get(self.url("/products"))
            .send()
            .await
            .context("Failed to send list products request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("List products failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<Vec<ProductResponse>>()
            .await
            .context("Failed to parse products response")?;

        Ok(result)
    }

    pub async fn get_product(&self, id: Uuid) -> Result<ProductResponse> {
        let resp = self
            .client
            .get(self.url(&format!("/products/{}", id)))
            .send()
            .await
            .context("Failed to send get product request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("Get product failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<ProductResponse>()
            .await
            .context("Failed to parse product response")?;

        Ok(result)
    }

    pub async fn create_activity(
        &self,
        product_id: Uuid,
        deposit_amount: u64,
        inflation_rate: u32,
        max_participants: u32,
        start_time: DateTime<Utc>,
        end_time: DateTime<Utc>,
        deadline_hours: Option<i64>,
    ) -> Result<ActivitySummaryResponse> {
        let req = CreateActivityRequest {
            product_id,
            deposit_amount,
            inflation_rate,
            max_participants,
            start_time,
            end_time,
            final_payment_deadline_hours: deadline_hours,
        };

        let resp = self
            .client
            .post(self.url("/activities"))
            .json(&req)
            .send()
            .await
            .context("Failed to send create activity request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("Create activity failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<ActivitySummaryResponse>()
            .await
            .context("Failed to parse activity response")?;

        Ok(result)
    }

    pub async fn list_activities(&self) -> Result<Vec<ActivitySummaryResponse>> {
        let resp = self
            .client
            .get(self.url("/activities"))
            .send()
            .await
            .context("Failed to send list activities request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("List activities failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<Vec<ActivitySummaryResponse>>()
            .await
            .context("Failed to parse activities response")?;

        Ok(result)
    }

    pub async fn get_activity(&self, id: Uuid) -> Result<ActivitySummaryResponse> {
        let resp = self
            .client
            .get(self.url(&format!("/activities/{}", id)))
            .send()
            .await
            .context("Failed to send get activity request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("Get activity failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<ActivitySummaryResponse>()
            .await
            .context("Failed to parse activity response")?;

        Ok(result)
    }

    pub async fn start_activity(&self, id: Uuid) -> Result<SimpleResponse> {
        let resp = self
            .client
            .put(self.url(&format!("/activities/{}/start", id)))
            .send()
            .await
            .context("Failed to send start activity request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("Start activity failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<SimpleResponse>()
            .await
            .context("Failed to parse response")?;

        Ok(result)
    }

    pub async fn end_activity(&self, id: Uuid) -> Result<SimpleResponse> {
        let resp = self
            .client
            .put(self.url(&format!("/activities/{}/end", id)))
            .send()
            .await
            .context("Failed to send end activity request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("End activity failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<SimpleResponse>()
            .await
            .context("Failed to parse response")?;

        Ok(result)
    }

    pub async fn cancel_activity(&self, id: Uuid) -> Result<SimpleResponse> {
        let resp = self
            .client
            .put(self.url(&format!("/activities/{}/cancel", id)))
            .send()
            .await
            .context("Failed to send cancel activity request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("Cancel activity failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<SimpleResponse>()
            .await
            .context("Failed to parse response")?;

        Ok(result)
    }

    pub async fn pay_deposit(&self, user_id: &str, activity_id: Uuid) -> Result<OrderDetailResponse> {
        let req = PayDepositRequest {
            user_id: user_id.to_string(),
            activity_id,
        };

        let resp = self
            .client
            .post(self.url("/orders/deposit"))
            .json(&req)
            .send()
            .await
            .context("Failed to send pay deposit request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("Pay deposit failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<OrderDetailResponse>()
            .await
            .context("Failed to parse order response")?;

        Ok(result)
    }

    pub async fn pay_final(&self, order_id: Uuid, amount: u64) -> Result<OrderDetailResponse> {
        let req = PayFinalRequest { paid_amount: amount };

        let resp = self
            .client
            .post(self.url(&format!("/orders/{}/final", order_id)))
            .json(&req)
            .send()
            .await
            .context("Failed to send pay final request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("Pay final failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<OrderDetailResponse>()
            .await
            .context("Failed to parse order response")?;

        Ok(result)
    }

    pub async fn get_order(&self, id: Uuid) -> Result<OrderDetailResponse> {
        let resp = self
            .client
            .get(self.url(&format!("/orders/{}", id)))
            .send()
            .await
            .context("Failed to send get order request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("Get order failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<OrderDetailResponse>()
            .await
            .context("Failed to parse order response")?;

        Ok(result)
    }

    pub async fn list_user_orders(&self, user_id: &str) -> Result<Vec<OrderDetailResponse>> {
        let resp = self
            .client
            .get(self.url(&format!("/users/{}/orders", user_id)))
            .send()
            .await
            .context("Failed to send list user orders request")?;

        if !resp.status().is_success() {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            anyhow::bail!("List user orders failed with status {}: {}", status, body);
        }

        let result = resp
            .json::<Vec<OrderDetailResponse>>()
            .await
            .context("Failed to parse orders response")?;

        Ok(result)
    }
}
