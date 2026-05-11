use std::collections::HashMap;
use std::sync::Arc;
use chrono::{Utc, NaiveDate, TimeZone};
use tokio::sync::Mutex;
use uuid::Uuid;

use crate::errors::{Result, TicketError};
use crate::models::*;
use crate::utils::*;

pub trait TicketRepository: Send + Sync {
    fn get_scenic(&self, scenic_id: &Uuid) -> Option<Scenic>;
    fn create_scenic(&mut self, scenic: Scenic);
    fn get_all_scenics(&self) -> Vec<Scenic>;
    
    fn get_order(&self, order_id: &Uuid) -> Option<Order>;
    fn create_order(&mut self, order: Order);
    fn update_order(&mut self, order: Order);
    
    fn get_daily_stats(&self, scenic_id: &Uuid, date: &NaiveDate) -> Option<ScenicDailyStats>;
    fn create_or_update_daily_stats(&mut self, stats: ScenicDailyStats);
}

pub struct InMemoryRepository {
    scenics: HashMap<Uuid, Scenic>,
    orders: HashMap<Uuid, Order>,
    daily_stats: HashMap<(Uuid, NaiveDate), ScenicDailyStats>,
}

impl InMemoryRepository {
    pub fn new() -> Self {
        Self {
            scenics: HashMap::new(),
            orders: HashMap::new(),
            daily_stats: HashMap::new(),
        }
    }
}

impl Default for InMemoryRepository {
    fn default() -> Self {
        Self::new()
    }
}

impl TicketRepository for InMemoryRepository {
    fn get_scenic(&self, scenic_id: &Uuid) -> Option<Scenic> {
        self.scenics.get(scenic_id).cloned()
    }

    fn create_scenic(&mut self, scenic: Scenic) {
        self.scenics.insert(scenic.id, scenic);
    }

    fn get_all_scenics(&self) -> Vec<Scenic> {
        self.scenics.values().cloned().collect()
    }

    fn get_order(&self, order_id: &Uuid) -> Option<Order> {
        self.orders.get(order_id).cloned()
    }

    fn create_order(&mut self, order: Order) {
        self.orders.insert(order.id, order);
    }

    fn update_order(&mut self, order: Order) {
        self.orders.insert(order.id, order);
    }

    fn get_daily_stats(&self, scenic_id: &Uuid, date: &NaiveDate) -> Option<ScenicDailyStats> {
        self.daily_stats.get(&(*scenic_id, *date)).cloned()
    }

    fn create_or_update_daily_stats(&mut self, stats: ScenicDailyStats) {
        self.daily_stats.insert((stats.scenic_id, stats.date), stats);
    }
}

pub struct TicketService<R: TicketRepository> {
    repository: Arc<Mutex<R>>,
}

impl<R: TicketRepository> TicketService<R> {
    pub fn new(repository: Arc<Mutex<R>>) -> Self {
        Self { repository }
    }
}

impl<R: TicketRepository> TicketService<R> {
    pub async fn create_scenic(&self, name: String, base_price: u32, daily_capacity: u32) -> Result<Scenic> {
        let now = Utc::now();
        let scenic = Scenic {
            id: Uuid::new_v4(),
            name,
            base_price,
            daily_capacity,
            created_at: now,
            updated_at: now,
        };
        
        let mut repo = self.repository.lock().await;
        repo.create_scenic(scenic.clone());
        Ok(scenic)
    }

    pub async fn get_all_scenics(&self) -> Result<Vec<Scenic>> {
        let repo = self.repository.lock().await;
        Ok(repo.get_all_scenics())
    }

    pub async fn get_scenic(&self, scenic_id: &Uuid) -> Result<Scenic> {
        let repo = self.repository.lock().await;
        repo.get_scenic(scenic_id).ok_or(TicketError::ScenicNotFound)
    }

    pub async fn create_order(&self, request: CreateOrderRequest) -> Result<CreateOrderResponse> {
        if request.tickets.is_empty() {
            return Err(TicketError::InvalidParam("至少需要购买一张门票".into()));
        }

        let mut repo = self.repository.lock().await;

        let scenic = repo.get_scenic(&request.scenic_id)
            .ok_or(TicketError::ScenicNotFound)?;

        let now = Utc::now();
        let today = now.date_naive();
        if request.use_date < today {
            return Err(TicketError::InvalidDate);
        }

        let season = get_season(&request.use_date);

        let adult_count = request.tickets.iter()
            .filter(|t| t.ticket_type == TicketType::Adult)
            .count();
        
        let free_child_requests: Vec<_> = request.tickets.iter()
            .filter(|t| {
                if t.ticket_type != TicketType::Child {
                    return false;
                }
                match t.child_height {
                    Some(h) => get_child_category(h) == ChildTicketCategory::Free,
                    None => false,
                }
            })
            .collect();

        if free_child_requests.len() > adult_count {
            return Err(TicketError::FreeChildrenExceeded);
        }

        let mut stats = repo.get_daily_stats(&request.scenic_id, &request.use_date)
            .unwrap_or_else(|| ScenicDailyStats {
                scenic_id: request.scenic_id,
                date: request.use_date,
                total_sold: 0,
                used_id_cards: Vec::new(),
            });

        let mut tickets_to_check_id: Vec<(String, String)> = Vec::new();
        for ticket_req in &request.tickets {
            let needs_id = match ticket_req.ticket_type {
                TicketType::Adult | TicketType::Elder | TicketType::Student => true,
                TicketType::Child => false,
            };

            if needs_id {
                let id_card = ticket_req.id_card.as_ref()
                    .ok_or_else(|| TicketError::InvalidParam(format!(
                        "{:?}票需要身份证", ticket_req.ticket_type
                    )))?;
                validate_id_card(id_card)?;

                if stats.used_id_cards.contains(id_card) {
                    return Err(TicketError::IdCardAlreadyPurchased);
                }

                if tickets_to_check_id.iter().any(|(id, _)| id == id_card) {
                    return Err(TicketError::IdCardAlreadyPurchased);
                }

                let name = ticket_req.name.as_ref()
                    .ok_or_else(|| TicketError::InvalidParam("需要姓名".into()))?;

                tickets_to_check_id.push((id_card.clone(), name.clone()));
            }
        }

        let new_count = request.tickets.len() as u32;
        if stats.total_sold + new_count > scenic.daily_capacity {
            return Err(TicketError::CapacityExceeded);
        }

        let mut ticket_items = Vec::new();
        let mut total_price = 0u32;

        for ticket_req in request.tickets {
            let (is_half_price, is_free, child_category, birth_date, is_elder_eligible) = 
                match ticket_req.ticket_type {
                    TicketType::Adult => {
                        (false, false, None, None, None)
                    }
                    TicketType::Child => {
                        let height = ticket_req.child_height
                            .ok_or_else(|| TicketError::InvalidParam("儿童票需要身高".into()))?;
                        let category = get_child_category(height);
                        let is_half = category == ChildTicketCategory::HalfPrice;
                        let is_free_child = category == ChildTicketCategory::Free;
                        (is_half, is_free_child, Some(category), None, None)
                    }
                    TicketType::Elder => {
                        let id_card = ticket_req.id_card.as_ref().unwrap();
                        let birth = extract_birth_date_from_id_card(id_card)?;
                        let eligible = is_elder(&birth, &request.use_date);
                        let is_half = eligible;
                        (is_half, false, None, Some(birth), Some(eligible))
                    }
                    TicketType::Student => {
                        if ticket_req.student_id.is_none() {
                            return Err(TicketError::InvalidParam("学生票需要学生证号".into()));
                        }
                        (true, false, None, None, None)
                    }
                };

            let actual_price = calculate_actual_price(
                scenic.base_price,
                season,
                is_half_price,
                is_free,
            );

            let item = TicketItem {
                id: Uuid::new_v4(),
                ticket_type: ticket_req.ticket_type,
                original_price: scenic.base_price,
                actual_price,
                id_card: ticket_req.id_card,
                name: ticket_req.name,
                child_height: ticket_req.child_height,
                child_category,
                birth_date,
                is_elder_eligible,
                student_id: ticket_req.student_id,
            };

            total_price += actual_price;
            ticket_items.push(item);
        }

        stats.total_sold += new_count;
        for (id_card, _) in tickets_to_check_id {
            stats.used_id_cards.push(id_card);
        }

        let order = Order {
            id: Uuid::new_v4(),
            scenic_id: request.scenic_id,
            use_date: request.use_date,
            tickets: ticket_items,
            total_price,
            created_at: now,
            is_refunded: false,
            refunded_at: None,
            refund_amount: None,
            refund_fee: None,
        };

        repo.create_order(order.clone());
        repo.create_or_update_daily_stats(stats.clone());

        Ok(CreateOrderResponse {
            order,
            use_count: stats.total_sold,
        })
    }

    pub async fn get_order(&self, order_id: &Uuid) -> Result<Order> {
        let repo = self.repository.lock().await;
        repo.get_order(order_id).ok_or(TicketError::OrderNotFound)
    }

    pub async fn refund_order(&self, order_id: &Uuid) -> Result<RefundResponse> {
        let mut repo = self.repository.lock().await;

        let mut order = repo.get_order(order_id)
            .ok_or(TicketError::OrderNotFound)?;

        if order.is_refunded {
            return Err(TicketError::OrderAlreadyRefunded);
        }

        let now = Utc::now();
        let refund_deadline = {
            let day_before = order.use_date.pred_opt().unwrap();
            Utc.from_utc_datetime(
                &day_before.and_hms_opt(23, 59, 59).unwrap()
            )
        };

        if now > refund_deadline {
            return Err(TicketError::RefundTimeExceeded);
        }

        let refund_fee = calculate_refund_fee(order.total_price);
        let refund_amount = order.total_price.saturating_sub(refund_fee);

        order.is_refunded = true;
        order.refunded_at = Some(now);
        order.refund_amount = Some(refund_amount);
        order.refund_fee = Some(refund_fee);

        let mut stats = repo.get_daily_stats(&order.scenic_id, &order.use_date)
            .ok_or_else(|| TicketError::Internal("日统计信息丢失".into()))?;

        let refund_count = order.tickets.len() as u32;
        stats.total_sold = stats.total_sold.saturating_sub(refund_count);

        for ticket in &order.tickets {
            if let Some(id_card) = &ticket.id_card {
                if let Some(idx) = stats.used_id_cards.iter().position(|id| id == id_card) {
                    stats.used_id_cards.remove(idx);
                }
            }
        }

        repo.update_order(order.clone());
        repo.create_or_update_daily_stats(stats);

        Ok(RefundResponse {
            order,
            refund_amount,
            refund_fee,
        })
    }

    pub async fn get_daily_stats(&self, scenic_id: &Uuid, date: &NaiveDate) -> Result<ScenicDailyStats> {
        let repo = self.repository.lock().await;
        let scenic = repo.get_scenic(scenic_id)
            .ok_or(TicketError::ScenicNotFound)?;
        
        Ok(repo.get_daily_stats(scenic_id, date)
            .unwrap_or_else(|| ScenicDailyStats {
                scenic_id: scenic.id,
                date: *date,
                total_sold: 0,
                used_id_cards: Vec::new(),
            }))
    }
}

pub type InMemoryTicketService = TicketService<InMemoryRepository>;

pub fn create_in_memory_service() -> Arc<InMemoryTicketService> {
    let repo = Arc::new(Mutex::new(InMemoryRepository::new()));
    Arc::new(TicketService::new(repo))
}
