use std::collections::{HashMap, HashSet};
use std::sync::{Arc, RwLock};

use chrono::{DateTime, Utc};
use rand::seq::IteratorRandom;
use uuid::Uuid;

use crate::errors::ProcurementError;
use crate::models::{
    AnonymousBid, Bid, Procurement, ProcurementStatus, QualificationLevel, RoundResult, Supplier,
};

pub struct ProcurementService {
    procurements: RwLock<HashMap<Uuid, Procurement>>,
    suppliers: RwLock<HashMap<Uuid, Supplier>>,
}

impl ProcurementService {
    pub fn new() -> Arc<Self> {
        Arc::new(Self {
            procurements: RwLock::new(HashMap::new()),
            suppliers: RwLock::new(HashMap::new()),
        })
    }

    pub fn create_supplier(&self, name: String, qualification: QualificationLevel) -> Supplier {
        let supplier = Supplier {
            id: Uuid::new_v4(),
            name,
            qualification,
        };
        let mut suppliers = self.suppliers.write().unwrap();
        suppliers.insert(supplier.id, supplier.clone());
        supplier
    }

    pub fn get_supplier(&self, id: Uuid) -> Option<Supplier> {
        let suppliers = self.suppliers.read().unwrap();
        suppliers.get(&id).cloned()
    }

    pub fn list_suppliers(&self) -> Vec<Supplier> {
        let suppliers = self.suppliers.read().unwrap();
        suppliers.values().cloned().collect()
    }

    pub fn create_procurement(
        &self,
        title: String,
        description: String,
        required_qualification: Option<QualificationLevel>,
    ) -> Procurement {
        let procurement = Procurement {
            id: Uuid::new_v4(),
            title,
            description,
            required_qualification,
            publish_date: Utc::now(),
            status: ProcurementStatus::Published,
            invited_suppliers: HashSet::new(),
            accepted_suppliers: HashSet::new(),
            current_round: 0,
            max_rounds: 3,
            bids: HashMap::new(),
            withdrawn_suppliers: HashSet::new(),
            winner_id: None,
            winner_bid: None,
        };
        let mut procurements = self.procurements.write().unwrap();
        procurements.insert(procurement.id, procurement.clone());
        procurement
    }

    pub fn get_procurement(&self, id: Uuid) -> Result<Procurement, ProcurementError> {
        let procurements = self.procurements.read().unwrap();
        procurements
            .get(&id)
            .cloned()
            .ok_or_else(|| ProcurementError::ProcurementNotFound(id.to_string()))
    }

    pub fn list_procurements(&self) -> Vec<Procurement> {
        let procurements = self.procurements.read().unwrap();
        procurements.values().cloned().collect()
    }

    pub fn invite_suppliers(
        &self,
        procurement_id: Uuid,
        supplier_ids: Vec<Uuid>,
    ) -> Result<Procurement, ProcurementError> {
        let mut procurements = self.procurements.write().unwrap();
        let procurement = procurements
            .get_mut(&procurement_id)
            .ok_or_else(|| ProcurementError::ProcurementNotFound(procurement_id.to_string()))?;

        if procurement.status != ProcurementStatus::Published {
            return Err(ProcurementError::InvalidStatusTransition {
                current: procurement.status.to_string(),
                operation: "邀请供应商".to_string(),
            });
        }

        let suppliers = self.suppliers.read().unwrap();
        for supplier_id in &supplier_ids {
            let supplier = suppliers
                .get(supplier_id)
                .ok_or_else(|| ProcurementError::SupplierNotFound(supplier_id.to_string()))?;

            if let Some(required) = procurement.required_qualification {
                if !supplier.qualification.meets_requirement(&required) {
                    return Err(ProcurementError::InsufficientQualification {
                        supplier: supplier.qualification.to_string(),
                        required: required.to_string(),
                    });
                }
            }
            procurement.invited_suppliers.insert(*supplier_id);
        }

        Ok(procurement.clone())
    }

    pub fn accept_invitation(
        &self,
        procurement_id: Uuid,
        supplier_id: Uuid,
    ) -> Result<Procurement, ProcurementError> {
        let mut procurements = self.procurements.write().unwrap();
        let procurement = procurements
            .get_mut(&procurement_id)
            .ok_or_else(|| ProcurementError::ProcurementNotFound(procurement_id.to_string()))?;

        if procurement.status != ProcurementStatus::Published {
            return Err(ProcurementError::InvalidStatusTransition {
                current: procurement.status.to_string(),
                operation: "接受邀请".to_string(),
            });
        }

        if !procurement.invited_suppliers.contains(&supplier_id) {
            return Err(ProcurementError::NotInvited);
        }

        procurement.accepted_suppliers.insert(supplier_id);
        Ok(procurement.clone())
    }

    pub fn start_bidding(&self, procurement_id: Uuid) -> Result<Procurement, ProcurementError> {
        let mut procurements = self.procurements.write().unwrap();
        let procurement = procurements
            .get_mut(&procurement_id)
            .ok_or_else(|| ProcurementError::ProcurementNotFound(procurement_id.to_string()))?;

        if procurement.status != ProcurementStatus::Published {
            return Err(ProcurementError::InvalidStatusTransition {
                current: procurement.status.to_string(),
                operation: "开始报价".to_string(),
            });
        }

        if procurement.accepted_suppliers.len() < 3 {
            procurement.status = ProcurementStatus::Cancelled;
            return Err(ProcurementError::InsufficientSuppliers);
        }

        procurement.status = ProcurementStatus::Bidding;
        procurement.current_round = 1;
        Ok(procurement.clone())
    }

    pub fn submit_bid(
        &self,
        procurement_id: Uuid,
        supplier_id: Uuid,
        price: f64,
        delivery_date: DateTime<Utc>,
    ) -> Result<Procurement, ProcurementError> {
        if price <= 0.0 {
            return Err(ProcurementError::InvalidPrice);
        }

        let mut procurements = self.procurements.write().unwrap();
        let procurement = procurements
            .get_mut(&procurement_id)
            .ok_or_else(|| ProcurementError::ProcurementNotFound(procurement_id.to_string()))?;

        if delivery_date < procurement.publish_date {
            return Err(ProcurementError::InvalidDeliveryDate);
        }

        if procurement.status != ProcurementStatus::Bidding {
            return Err(ProcurementError::NotInBiddingPhase);
        }

        if !procurement.accepted_suppliers.contains(&supplier_id) {
            return Err(ProcurementError::NotInvited);
        }

        if procurement.withdrawn_suppliers.contains(&supplier_id) {
            return Err(ProcurementError::SupplierWithdrawn);
        }

        let current_round = procurement.current_round;

        if procurement
            .bids
            .get(&current_round)
            .map(|b| b.contains_key(&supplier_id))
            .unwrap_or(false)
        {
            return Err(ProcurementError::AlreadySubmitted);
        }

        if current_round > 1 {
            if let Some(prev_round_bids) = procurement.bids.get(&(current_round - 1)) {
                if let Some(prev_bid) = prev_round_bids.get(&supplier_id) {
                    if price > prev_bid.price {
                        return Err(ProcurementError::PriceNotDecreased);
                    }
                }
            }
        }

        let bid = Bid {
            supplier_id,
            procurement_id,
            round: current_round,
            price,
            delivery_date,
            submitted_at: Utc::now(),
        };

        let round_bids = procurement.bids.entry(current_round).or_insert_with(HashMap::new);
        round_bids.insert(supplier_id, bid);

        Ok(procurement.clone())
    }

    pub fn withdraw_from_bidding(
        &self,
        procurement_id: Uuid,
        supplier_id: Uuid,
    ) -> Result<Procurement, ProcurementError> {
        let mut procurements = self.procurements.write().unwrap();
        let procurement = procurements
            .get_mut(&procurement_id)
            .ok_or_else(|| ProcurementError::ProcurementNotFound(procurement_id.to_string()))?;

        if !procurement.accepted_suppliers.contains(&supplier_id) {
            return Err(ProcurementError::NotInvited);
        }

        procurement.withdrawn_suppliers.insert(supplier_id);
        Ok(procurement.clone())
    }

    pub fn get_round_result(&self, procurement_id: Uuid, round: u32) -> Option<RoundResult> {
        let procurements = self.procurements.read().unwrap();
        let procurement = procurements.get(&procurement_id)?;

        if round > procurement.current_round || round == 0 {
            return None;
        }

        let round_bids = procurement.bids.get(&round)?;
        let mut lowest_price: Option<f64> = None;

        let anonymous_bids: Vec<AnonymousBid> = round_bids
            .values()
            .map(|bid| {
                if lowest_price.map_or(true, |p| bid.price < p) {
                    lowest_price = Some(bid.price);
                }
                AnonymousBid {
                    price: bid.price,
                    delivery_date: bid.delivery_date,
                }
            })
            .collect();

        Some(RoundResult {
            round,
            bids: anonymous_bids,
            lowest_price,
        })
    }

    pub fn advance_round(&self, procurement_id: Uuid) -> Result<Procurement, ProcurementError> {
        let mut procurements = self.procurements.write().unwrap();
        let procurement = procurements
            .get_mut(&procurement_id)
            .ok_or_else(|| ProcurementError::ProcurementNotFound(procurement_id.to_string()))?;

        if procurement.status != ProcurementStatus::Bidding {
            return Err(ProcurementError::NotInBiddingPhase);
        }

        if procurement.current_round >= procurement.max_rounds {
            return self.evaluate_winner_internal(procurement);
        }

        let active_suppliers_count = procurement
            .accepted_suppliers
            .difference(&procurement.withdrawn_suppliers)
            .count();

        if active_suppliers_count == 0 {
            procurement.status = ProcurementStatus::Cancelled;
            return Err(ProcurementError::NoValidBids);
        }

        procurement.current_round += 1;
        Ok(procurement.clone())
    }

    pub fn evaluate_winner(&self, procurement_id: Uuid) -> Result<Procurement, ProcurementError> {
        let mut procurements = self.procurements.write().unwrap();
        let procurement = procurements
            .get_mut(&procurement_id)
            .ok_or_else(|| ProcurementError::ProcurementNotFound(procurement_id.to_string()))?;

        self.evaluate_winner_internal(procurement)
    }

    fn evaluate_winner_internal(
        &self,
        procurement: &mut Procurement,
    ) -> Result<Procurement, ProcurementError> {
        procurement.status = ProcurementStatus::Evaluating;

        let final_round = procurement.current_round;
        let round_bids = procurement
            .bids
            .get(&final_round)
            .ok_or(ProcurementError::NoValidBids)?;

        if round_bids.is_empty() {
            procurement.status = ProcurementStatus::Cancelled;
            return Err(ProcurementError::NoValidBids);
        }

        let mut lowest_price = f64::INFINITY;
        for bid in round_bids.values() {
            if bid.price < lowest_price {
                lowest_price = bid.price;
            }
        }

        let lowest_bidders: Vec<&Bid> = round_bids
            .values()
            .filter(|bid| bid.price == lowest_price)
            .collect();

        let winner = if lowest_bidders.len() == 1 {
            lowest_bidders[0]
        } else {
            let mut rng = rand::thread_rng();
            lowest_bidders
                .into_iter()
                .choose(&mut rng)
                .ok_or(ProcurementError::NoValidBids)?
        };

        procurement.winner_id = Some(winner.supplier_id);
        procurement.winner_bid = Some(winner.clone());
        procurement.status = ProcurementStatus::Awarded;

        Ok(procurement.clone())
    }

    pub fn update_procurement(
        &self,
        procurement_id: Uuid,
        title: Option<String>,
        description: Option<String>,
        required_qualification: Option<QualificationLevel>,
    ) -> Result<Procurement, ProcurementError> {
        let mut procurements = self.procurements.write().unwrap();
        let procurement = procurements
            .get_mut(&procurement_id)
            .ok_or_else(|| ProcurementError::ProcurementNotFound(procurement_id.to_string()))?;

        if procurement.status != ProcurementStatus::Published {
            return Err(ProcurementError::CannotModifyDuringBidding);
        }

        if let Some(t) = title {
            procurement.title = t;
        }
        if let Some(d) = description {
            procurement.description = d;
        }
        if required_qualification.is_some() {
            procurement.required_qualification = required_qualification;
        }

        Ok(procurement.clone())
    }
}

impl Default for ProcurementService {
    fn default() -> Self {
        Self {
            procurements: RwLock::new(HashMap::new()),
            suppliers: RwLock::new(HashMap::new()),
        }
    }
}
