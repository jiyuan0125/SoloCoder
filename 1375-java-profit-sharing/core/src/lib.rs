use chrono::{Datelike, Duration, NaiveDate};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OrderType {
    Standard,
    Premium,
    VIP,
}

impl OrderType {
    pub fn get_share_ratios(&self) -> ShareRatios {
        match self {
            OrderType::Standard => ShareRatios {
                platform: 0.10,
                merchant: 0.70,
                promoter: 0.20,
            },
            OrderType::Premium => ShareRatios {
                platform: 0.15,
                merchant: 0.60,
                promoter: 0.25,
            },
            OrderType::VIP => ShareRatios {
                platform: 0.20,
                merchant: 0.50,
                promoter: 0.30,
            },
        }
    }
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
pub struct ShareRatios {
    pub platform: f64,
    pub merchant: f64,
    pub promoter: f64,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OrderStatus {
    Pending,
    Completed,
    Settled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: Uuid,
    pub order_type: OrderType,
    pub amount: u64,
    pub merchant_id: Uuid,
    pub promoter_id: Option<Uuid>,
    pub status: OrderStatus,
    pub completed_date: Option<NaiveDate>,
    pub settlement_id: Option<Uuid>,
    pub created_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Promoter {
    pub id: Uuid,
    pub name: String,
    pub has_payment_account: bool,
    pub pending_amount: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SettlementShare {
    pub platform_amount: u64,
    pub merchant_amount: u64,
    pub promoter_amount: u64,
    pub promoter_pending: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderSettlement {
    pub order_id: Uuid,
    pub share: SettlementShare,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Settlement {
    pub id: Uuid,
    pub settlement_date: NaiveDate,
    pub orders: Vec<OrderSettlement>,
    pub total_platform: u64,
    pub total_merchant: u64,
    pub total_promoter: u64,
    pub total_pending: u64,
    pub created_at: chrono::DateTime<chrono::Utc>,
}

pub struct ProfitSharingService {
    orders: HashMap<Uuid, Order>,
    promoters: HashMap<Uuid, Promoter>,
    settlements: HashMap<Uuid, Settlement>,
}

impl ProfitSharingService {
    pub fn new() -> Self {
        Self {
            orders: HashMap::new(),
            promoters: HashMap::new(),
            settlements: HashMap::new(),
        }
    }

    pub fn create_order(
        &mut self,
        order_type: OrderType,
        amount: u64,
        merchant_id: Uuid,
        promoter_id: Option<Uuid>,
    ) -> Order {
        let order = Order {
            id: Uuid::new_v4(),
            order_type,
            amount,
            merchant_id,
            promoter_id,
            status: OrderStatus::Pending,
            completed_date: None,
            settlement_id: None,
            created_at: chrono::Utc::now(),
        };
        self.orders.insert(order.id, order.clone());
        order
    }

    pub fn complete_order(&mut self, order_id: Uuid, completed_date: NaiveDate) -> Result<Order, String> {
        let order = self.orders.get_mut(&order_id)
            .ok_or_else(|| format!("Order not found: {}", order_id))?;
        
        if order.status != OrderStatus::Pending {
            return Err(format!("Order {} is not pending", order_id));
        }
        
        order.status = OrderStatus::Completed;
        order.completed_date = Some(completed_date);
        Ok(order.clone())
    }

    pub fn register_promoter(&mut self, name: String) -> Promoter {
        let promoter = Promoter {
            id: Uuid::new_v4(),
            name,
            has_payment_account: false,
            pending_amount: 0,
        };
        self.promoters.insert(promoter.id, promoter.clone());
        promoter
    }

    pub fn bind_promoter_account(&mut self, promoter_id: Uuid) -> Result<u64, String> {
        let promoter = self.promoters.get_mut(&promoter_id)
            .ok_or_else(|| format!("Promoter not found: {}", promoter_id))?;
        
        let pending_amount = promoter.pending_amount;
        promoter.has_payment_account = true;
        promoter.pending_amount = 0;
        
        Ok(pending_amount)
    }

    pub fn calculate_share(&self, order: &Order) -> SettlementShare {
        let ratios = order.order_type.get_share_ratios();
        let total = order.amount as f64;
        
        let platform_raw = total * ratios.platform;
        let promoter_raw = total * ratios.promoter;
        
        let platform_amount = platform_raw.round() as u64;
        let promoter_amount = if order.promoter_id.is_some() {
            promoter_raw.round() as u64
        } else {
            0
        };
        
        let merchant_amount = order.amount
            .checked_sub(platform_amount)
            .and_then(|v| v.checked_sub(promoter_amount))
            .unwrap_or(0);
        
        let promoter_pending = if let Some(promoter_id) = order.promoter_id {
            self.promoters.get(&promoter_id)
                .map(|p| !p.has_payment_account)
                .unwrap_or(false)
        } else {
            false
        };
        
        SettlementShare {
            platform_amount,
            merchant_amount,
            promoter_amount,
            promoter_pending,
        }
    }

    pub fn get_settlement_date(&self, completed_date: NaiveDate) -> NaiveDate {
        let mut settlement_date = completed_date + Duration::days(1);
        while !self.is_workday(settlement_date) {
            settlement_date += Duration::days(1);
        }
        settlement_date
    }

    pub fn is_workday(&self, date: NaiveDate) -> bool {
        let weekday = date.weekday();
        weekday != chrono::Weekday::Sat && weekday != chrono::Weekday::Sun
    }

    pub fn process_settlement(&mut self, settlement_date: NaiveDate) -> Result<Settlement, String> {
        let orders_to_settle: Vec<Uuid> = self.orders
            .iter()
            .filter(|(_, order)| {
                order.status == OrderStatus::Completed
                    && order.settlement_id.is_none()
                    && order.completed_date
                        .map(|cd| self.get_settlement_date(cd) == settlement_date)
                        .unwrap_or(false)
            })
            .map(|(id, _)| *id)
            .collect();

        let mut order_settlements = Vec::new();
        let mut total_platform = 0u64;
        let mut total_merchant = 0u64;
        let mut total_promoter = 0u64;
        let mut total_pending = 0u64;

        for order_id in &orders_to_settle {
            let order = self.orders.get(order_id).unwrap();
            let share = self.calculate_share(order);
            
            if share.promoter_pending {
                if let Some(promoter_id) = order.promoter_id {
                    if let Some(promoter) = self.promoters.get_mut(&promoter_id) {
                        promoter.pending_amount += share.promoter_amount;
                    }
                }
                total_pending += share.promoter_amount;
            } else {
                total_promoter += share.promoter_amount;
            }
            
            total_platform += share.platform_amount;
            total_merchant += share.merchant_amount;
            
            order_settlements.push(OrderSettlement {
                order_id: *order_id,
                share: share.clone(),
            });
        }

        let settlement = Settlement {
            id: Uuid::new_v4(),
            settlement_date,
            orders: order_settlements,
            total_platform,
            total_merchant,
            total_promoter,
            total_pending,
            created_at: chrono::Utc::now(),
        };

        for order_id in &orders_to_settle {
            if let Some(order) = self.orders.get_mut(order_id) {
                order.status = OrderStatus::Settled;
                order.settlement_id = Some(settlement.id);
            }
        }

        self.settlements.insert(settlement.id, settlement.clone());
        Ok(settlement)
    }

    pub fn get_order(&self, order_id: Uuid) -> Option<Order> {
        self.orders.get(&order_id).cloned()
    }

    pub fn get_all_orders(&self) -> Vec<Order> {
        self.orders.values().cloned().collect()
    }

    pub fn get_promoter(&self, promoter_id: Uuid) -> Option<Promoter> {
        self.promoters.get(&promoter_id).cloned()
    }

    pub fn get_all_promoters(&self) -> Vec<Promoter> {
        self.promoters.values().cloned().collect()
    }

    pub fn get_settlement(&self, settlement_id: Uuid) -> Option<Settlement> {
        self.settlements.get(&settlement_id).cloned()
    }

    pub fn get_all_settlements(&self) -> Vec<Settlement> {
        self.settlements.values().cloned().collect()
    }
}

impl Default for ProfitSharingService {
    fn default() -> Self {
        Self::new()
    }
}
