use crate::errors::{Result, StoredValueError};
use crate::models::{ConsumeBreakdown, Member, MemberLevel, RechargeConfig, Transaction, TransactionType};
use crate::storage::InMemoryStorage;
use std::sync::Arc;
use tokio::sync::Mutex;

pub struct StoredValueService {
    storage: Arc<InMemoryStorage>,
    recharge_config: RechargeConfig,
    locks: Arc<Mutex<std::collections::HashMap<String, ()>>>,
}

impl StoredValueService {
    pub fn new(storage: Arc<InMemoryStorage>) -> Self {
        Self {
            storage,
            recharge_config: RechargeConfig::default(),
            locks: Arc::new(Mutex::new(std::collections::HashMap::new())),
        }
    }
    
    pub async fn create_member(&self, name: String, level: MemberLevel) -> Member {
        let member = Member::new(name, level);
        self.storage.create_member(member.clone()).await;
        member
    }
    
    pub async fn get_member(&self, id: &str) -> Result<Member> {
        self.storage
            .get_member(id)
            .await
            .ok_or_else(|| StoredValueError::MemberNotFound(id.to_string()))
    }
    
    pub async fn list_members(&self) -> Vec<Member> {
        self.storage.list_members().await
    }
    
    pub async fn recharge(&self, member_id: &str, amount: u64) -> Result<(Member, Vec<Transaction>)> {
        if amount == 0 {
            return Err(StoredValueError::InvalidRechargeAmount);
        }
        
        let _guard = self.acquire_lock(member_id).await;
        
        let mut member = self.get_member(member_id).await?;
        
        let gift_amount = self.recharge_config.get_gift_amount(amount);
        
        member.principal_balance += amount;
        if gift_amount > 0 {
            member.gift_balance += gift_amount;
        }
        member.updated_at = chrono::Utc::now();
        
        self.storage.update_member(member.clone()).await;
        
        let mut transactions = Vec::new();
        
        let recharge_txn = Transaction::new(
            member_id.to_string(),
            TransactionType::Recharge,
            amount,
            amount,
            amount as i64,
            0,
            0,
            format!("充值 {} 元", amount),
            None,
        );
        transactions.push(recharge_txn.clone());
        self.storage.add_transaction(recharge_txn).await;
        
        if gift_amount > 0 {
            let gift_txn = Transaction::new(
                member_id.to_string(),
                TransactionType::Gift,
                gift_amount,
                gift_amount,
                0,
                gift_amount as i64,
                0,
                format!("充值赠送 {} 元", gift_amount),
                None,
            );
            transactions.push(gift_txn.clone());
            self.storage.add_transaction(gift_txn).await;
        }
        
        Ok((member, transactions))
    }
    
    pub async fn consume(&self, member_id: &str, original_price: u64, description: String) -> Result<(Member, Transaction, ConsumeBreakdown)> {
        if original_price == 0 {
            return Err(StoredValueError::InvalidConsumeAmount);
        }
        
        let _guard = self.acquire_lock(member_id).await;
        
        let mut member = self.get_member(member_id).await?;
        
        let discount = member.level.discount();
        let discounted_price = (original_price as f64 * discount) as u64;
        let points_earned = original_price;
        
        if member.total_balance() < discounted_price {
            return Err(StoredValueError::InsufficientBalance(
                discounted_price,
                member.total_balance(),
            ));
        }
        
        let breakdown = self.calculate_deduction(&member, discounted_price);
        
        member.gift_balance -= breakdown.gift_deducted;
        member.principal_balance -= breakdown.principal_deducted;
        member.points += points_earned;
        member.updated_at = chrono::Utc::now();
        
        self.storage.update_member(member.clone()).await;
        
        let transaction = Transaction::new(
            member_id.to_string(),
            TransactionType::Consume,
            original_price,
            discounted_price,
            -(breakdown.principal_deducted as i64),
            -(breakdown.gift_deducted as i64),
            points_earned as i64,
            description,
            None,
        );
        self.storage.add_transaction(transaction.clone()).await;
        
        Ok((member, transaction, breakdown))
    }
    
    pub async fn refund(&self, consume_transaction_id: &str) -> Result<(Member, Transaction)> {
        let _txn = self.storage
            .get_transaction(consume_transaction_id)
            .await
            .ok_or_else(|| StoredValueError::TransactionNotFound(consume_transaction_id.to_string()))?;
        
        if _txn.transaction_type != TransactionType::Consume {
            return Err(StoredValueError::InternalError("只能对消费交易进行退款".to_string()));
        }
        
        let member_id = _txn.member_id.clone();
        let _guard = self.acquire_lock(&member_id).await;
        
        let mut member = self.get_member(&member_id).await?;
        
        let original_consume = _txn.discounted_price;
        let principal_deducted = -_txn.principal_amount as u64;
        let gift_deducted = -_txn.gift_amount as u64;
        let points_earned = _txn.points as u64;
        
        member.principal_balance += principal_deducted;
        member.gift_balance += gift_deducted;
        member.points = member.points.saturating_sub(points_earned);
        member.updated_at = chrono::Utc::now();
        
        self.storage.update_member(member.clone()).await;
        
        let refund_txn = Transaction::new(
            member_id,
            TransactionType::Refund,
            original_consume,
            original_consume,
            principal_deducted as i64,
            gift_deducted as i64,
            -(points_earned as i64),
            format!("退款，关联消费: {}", consume_transaction_id),
            Some(consume_transaction_id.to_string()),
        );
        self.storage.add_transaction(refund_txn.clone()).await;
        
        Ok((member, refund_txn))
    }
    
    pub async fn list_member_transactions(&self, member_id: &str) -> Result<Vec<Transaction>> {
        self.get_member(member_id).await?;
        Ok(self.storage.list_member_transactions(member_id).await)
    }
    
    fn calculate_deduction(&self, member: &Member, amount: u64) -> ConsumeBreakdown {
        let gift_deducted = std::cmp::min(member.gift_balance, amount);
        let principal_deducted = amount - gift_deducted;
        
        ConsumeBreakdown {
            gift_deducted,
            principal_deducted,
        }
    }
    
    async fn acquire_lock(&self, member_id: &str) -> MemberLockGuard {
        let mut locks = self.locks.lock().await;
        while locks.contains_key(member_id) {
            drop(locks);
            tokio::time::sleep(tokio::time::Duration::from_millis(10)).await;
            locks = self.locks.lock().await;
        }
        locks.insert(member_id.to_string(), ());
        MemberLockGuard {
            member_id: member_id.to_string(),
            locks: Arc::clone(&self.locks),
        }
    }
}

struct MemberLockGuard {
    member_id: String,
    locks: Arc<Mutex<std::collections::HashMap<String, ()>>>,
}

impl Drop for MemberLockGuard {
    fn drop(&mut self) {
        let member_id = self.member_id.clone();
        let locks = Arc::clone(&self.locks);
        tokio::spawn(async move {
            let mut locks = locks.lock().await;
            locks.remove(&member_id);
        });
    }
}
