use std::collections::HashMap;
use tokio::sync::RwLock;
use crate::models::{Member, Transaction};

pub struct InMemoryStorage {
    members: RwLock<HashMap<String, Member>>,
    transactions: RwLock<HashMap<String, Transaction>>,
    member_transactions: RwLock<HashMap<String, Vec<String>>>,
}

impl InMemoryStorage {
    pub fn new() -> Self {
        Self {
            members: RwLock::new(HashMap::new()),
            transactions: RwLock::new(HashMap::new()),
            member_transactions: RwLock::new(HashMap::new()),
        }
    }
    
    pub async fn create_member(&self, member: Member) {
        let mut members = self.members.write().await;
        let mut member_txns = self.member_transactions.write().await;
        let member_id = member.id.clone();
        members.insert(member.id.clone(), member);
        member_txns.insert(member_id, Vec::new());
    }
    
    pub async fn get_member(&self, id: &str) -> Option<Member> {
        let members = self.members.read().await;
        members.get(id).cloned()
    }
    
    pub async fn list_members(&self) -> Vec<Member> {
        let members = self.members.read().await;
        members.values().cloned().collect()
    }
    
    pub async fn update_member(&self, member: Member) {
        let mut members = self.members.write().await;
        members.insert(member.id.clone(), member);
    }
    
    pub async fn add_transaction(&self, transaction: Transaction) {
        let mut transactions = self.transactions.write().await;
        let mut member_txns = self.member_transactions.write().await;
        let txn_id = transaction.id.clone();
        let member_id = transaction.member_id.clone();
        transactions.insert(txn_id.clone(), transaction);
        if let Some(txns) = member_txns.get_mut(&member_id) {
            txns.push(txn_id);
        }
    }
    
    pub async fn get_transaction(&self, id: &str) -> Option<Transaction> {
        let transactions = self.transactions.read().await;
        transactions.get(id).cloned()
    }
    
    pub async fn list_member_transactions(&self, member_id: &str) -> Vec<Transaction> {
        let member_txns = self.member_transactions.read().await;
        let transactions = self.transactions.read().await;
        member_txns
            .get(member_id)
            .map(|ids| {
                ids.iter()
                    .filter_map(|id| transactions.get(id).cloned())
                    .collect()
            })
            .unwrap_or_default()
    }
}

impl Default for InMemoryStorage {
    fn default() -> Self {
        Self::new()
    }
}
