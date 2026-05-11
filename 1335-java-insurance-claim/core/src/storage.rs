use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use crate::models::{Claim, Policy};

#[derive(Debug, Clone, Default)]
pub struct InMemoryStorage {
    claims: Arc<RwLock<HashMap<String, Claim>>>,
    policies: Arc<RwLock<HashMap<String, Policy>>>,
}

impl InMemoryStorage {
    pub fn new() -> Self {
        Self::default()
    }

    pub async fn insert_claim(&self, claim: Claim) {
        let mut claims = self.claims.write().await;
        claims.insert(claim.id.clone(), claim);
    }

    pub async fn update_claim(&self, claim: Claim) {
        let mut claims = self.claims.write().await;
        claims.insert(claim.id.clone(), claim);
    }

    pub async fn get_claim(&self, id: &str) -> Option<Claim> {
        let claims = self.claims.read().await;
        claims.get(id).cloned()
    }

    pub async fn get_all_claims(&self) -> Vec<Claim> {
        let claims = self.claims.read().await;
        claims.values().cloned().collect()
    }

    pub async fn insert_policy(&self, policy: Policy) {
        let mut policies = self.policies.write().await;
        policies.insert(policy.id.clone(), policy);
    }

    pub async fn get_policy(&self, id: &str) -> Option<Policy> {
        let policies = self.policies.read().await;
        policies.get(id).cloned()
    }

    pub async fn get_all_policies(&self) -> Vec<Policy> {
        let policies = self.policies.read().await;
        policies.values().cloned().collect()
    }
}
