use std::collections::HashMap;
use std::sync::Mutex;
use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::error::*;
use crate::models::*;
use crate::types::*;

const EXPIRY_SKIP_DAYS: i64 = 90;
const EXPIRY_ALERT_DAYS: i64 = 30;

pub struct DrugTraceService {
    batches: Mutex<HashMap<Uuid, DrugBatch>>,
    inbound_records: Mutex<Vec<InboundRecord>>,
    outbound_records: Mutex<Vec<OutboundRecord>>,
    batch_no_to_id: Mutex<HashMap<String, Uuid>>,
}

impl Default for DrugTraceService {
    fn default() -> Self {
        Self::new()
    }
}

impl DrugTraceService {
    pub fn new() -> Self {
        Self {
            batches: Mutex::new(HashMap::new()),
            inbound_records: Mutex::new(Vec::new()),
            outbound_records: Mutex::new(Vec::new()),
            batch_no_to_id: Mutex::new(HashMap::new()),
        }
    }

    fn validate_inbound(&self, req: &InboundRequest) -> Result<()> {
        if req.drug_name.trim().is_empty() {
            return Err(DrugTraceError::EmptyDrugName);
        }
        if req.batch_no.trim().is_empty() {
            return Err(DrugTraceError::EmptyBatchNo);
        }
        if req.supplier.trim().is_empty() {
            return Err(DrugTraceError::EmptySupplier);
        }
        if req.quantity == 0 {
            return Err(DrugTraceError::InvalidQuantity);
        }
        if req.expiry_date <= req.production_date {
            return Err(DrugTraceError::InvalidDate);
        }
        Ok(())
    }

    pub fn inbound(&self, req: InboundRequest) -> Result<DrugBatch> {
        self.validate_inbound(&req)?;

        let now = Utc::now();
        let batch_id = Uuid::new_v4();

        let batch = DrugBatch {
            id: batch_id,
            drug_name: req.drug_name.clone(),
            batch_no: req.batch_no.clone(),
            production_date: req.production_date,
            expiry_date: req.expiry_date,
            supplier: req.supplier.clone(),
            total_quantity: req.quantity,
            remaining_quantity: req.quantity,
            status: BatchStatus::InStock,
            created_at: now,
            recall_flag: false,
        };

        let inbound = InboundRecord {
            id: Uuid::new_v4(),
            batch_id,
            drug_name: req.drug_name.clone(),
            batch_no: req.batch_no.clone(),
            quantity: req.quantity,
            supplier: req.supplier.clone(),
            created_at: now,
        };

        {
            let mut batches = self.batches.lock().unwrap();
            batches.insert(batch_id, batch.clone());
        }
        {
            let mut inbound_records = self.inbound_records.lock().unwrap();
            inbound_records.push(inbound);
        }
        {
            let mut batch_no_map = self.batch_no_to_id.lock().unwrap();
            batch_no_map.insert(req.batch_no.clone(), batch_id);
        }

        Ok(batch)
    }

    fn days_until_expiry(expiry: DateTime<Utc>) -> i64 {
        (expiry - Utc::now()).num_days()
    }

    fn is_expired(expiry: DateTime<Utc>) -> bool {
        Self::days_until_expiry(expiry) < 0
    }

    fn is_near_expiry(expiry: DateTime<Utc>, days: i64) -> bool {
        let remaining = Self::days_until_expiry(expiry);
        remaining >= 0 && remaining < days
    }

    fn get_available_batches_for_drug(&self, drug_name: &str) -> Vec<DrugBatch> {
        let batches = self.batches.lock().unwrap();
        batches
            .values()
            .filter(|b| {
                b.drug_name == drug_name
                    && b.remaining_quantity > 0
                    && b.status != BatchStatus::Frozen
                    && b.status != BatchStatus::Recalled
                    && !b.recall_flag
            })
            .cloned()
            .collect()
    }

    fn select_batches_for_outbound(
        &self,
        mut batches: Vec<DrugBatch>,
        quantity: u32,
    ) -> (Vec<(DrugBatch, u32)>, bool, Option<String>) {
        let mut result = Vec::new();
        let mut remaining = quantity;
        let mut has_warning = false;
        let mut warning_message = None;

        batches.sort_by(|a, b| {
            let a_expired = Self::is_expired(a.expiry_date);
            let b_expired = Self::is_expired(b.expiry_date);
            if a_expired && !b_expired {
                return std::cmp::Ordering::Greater;
            }
            if !a_expired && b_expired {
                return std::cmp::Ordering::Less;
            }
            a.expiry_date.cmp(&b.expiry_date)
        });

        let normal_batches: Vec<DrugBatch> = batches
            .iter()
            .filter(|b| !Self::is_near_expiry(b.expiry_date, EXPIRY_SKIP_DAYS))
            .cloned()
            .collect();

        let near_expiry_batches: Vec<DrugBatch> = batches
            .iter()
            .filter(|b| {
                let days = Self::days_until_expiry(b.expiry_date);
                days >= 0 && days < EXPIRY_SKIP_DAYS
            })
            .cloned()
            .collect();

        let mut use_normal = !normal_batches.is_empty();

        if !use_normal {
            use_normal = true;
            has_warning = true;
            warning_message = Some(
                "所有批次均在90天内过期，将出库最早过期的批次".to_string(),
            );
        }

        let priority_batches = if use_normal { &normal_batches } else { &near_expiry_batches };

        for batch in priority_batches {
            if remaining == 0 {
                break;
            }
            if Self::is_expired(batch.expiry_date) {
                continue;
            }
            if batch.remaining_quantity == 0 {
                continue;
            }

            let take = std::cmp::min(batch.remaining_quantity, remaining);
            result.push((batch.clone(), take));
            remaining -= take;
        }

        if remaining > 0 {
            for batch in &near_expiry_batches {
                if remaining == 0 {
                    break;
                }
                if Self::is_expired(batch.expiry_date) {
                    continue;
                }
                if batch.remaining_quantity == 0 {
                    continue;
                }
                if result.iter().any(|(b, _)| b.id == batch.id) {
                    continue;
                }

                let take = std::cmp::min(batch.remaining_quantity, remaining);
                result.push((batch.clone(), take));
                remaining -= take;
                has_warning = true;
                warning_message = Some(format!(
                    "库存不足，使用了临近过期批次({}天内)",
                    EXPIRY_SKIP_DAYS
                ));
            }
        }

        (result, has_warning, warning_message)
    }

    pub fn outbound(&self, req: OutboundRequest) -> Result<OutboundResult> {
        if req.quantity == 0 {
            return Err(DrugTraceError::InvalidQuantity);
        }
        if req.drug_name.trim().is_empty() {
            return Err(DrugTraceError::EmptyDrugName);
        }
        if req.customer.trim().is_empty() {
            return Err(DrugTraceError::EmptySupplier);
        }

        let available = self.get_available_batches_for_drug(&req.drug_name);
        if available.is_empty() {
            return Err(DrugTraceError::DrugNotFound(req.drug_name.clone()));
        }

        let total_available: u32 = available
            .iter()
            .filter(|b| !Self::is_expired(b.expiry_date))
            .map(|b| b.remaining_quantity)
            .sum();

        if total_available < req.quantity {
            return Err(DrugTraceError::InsufficientStock);
        }

        let (selected_batches, has_warning, warning_message) =
            self.select_batches_for_outbound(available, req.quantity);

        if selected_batches.is_empty() {
            return Err(DrugTraceError::InsufficientStock);
        }

        let details: Vec<OutboundDetail> = selected_batches
            .iter()
            .map(|(batch, qty)| OutboundDetail {
                batch_id: batch.id,
                batch_no: batch.batch_no.clone(),
                quantity: *qty,
            })
            .collect();

        let total_quantity: u32 = selected_batches.iter().map(|(_, qty)| *qty).sum();

        {
            let mut batches = self.batches.lock().unwrap();
            for (batch, take) in &selected_batches {
                if let Some(existing) = batches.get_mut(&batch.id) {
                    existing.remaining_quantity -= take;
                    if existing.remaining_quantity == 0 {
                        existing.status = BatchStatus::FullyOut;
                    } else {
                        existing.status = BatchStatus::PartiallyOut;
                    }
                }
            }
        }

        let now = Utc::now();
        let record = OutboundRecord {
            id: Uuid::new_v4(),
            drug_name: req.drug_name.clone(),
            customer: req.customer.clone(),
            details,
            total_quantity,
            has_warning,
            warning_message,
            created_at: now,
        };

        {
            let mut outbound_records = self.outbound_records.lock().unwrap();
            outbound_records.push(record.clone());
        }

        let alerts = self.get_expiry_alerts_internal();

        Ok(OutboundResult { record, alerts })
    }

    fn get_expiry_alerts_internal(&self) -> Vec<ExpiryAlert> {
        let batches = self.batches.lock().unwrap();
        let now = Utc::now();

        batches
            .values()
            .filter(|b| b.remaining_quantity > 0 && b.status != BatchStatus::Frozen)
            .filter_map(|b| {
                let remaining_days = (b.expiry_date - now).num_days();
                let is_expired = remaining_days < 0;
                let is_near_expiry = remaining_days >= 0 && remaining_days < EXPIRY_ALERT_DAYS;

                if is_expired || is_near_expiry {
                    Some(ExpiryAlert {
                        batch_no: b.batch_no.clone(),
                        drug_name: b.drug_name.clone(),
                        remaining_days,
                        remaining_quantity: b.remaining_quantity,
                        is_expired,
                    })
                } else {
                    None
                }
            })
            .collect()
    }

    pub fn get_expiry_alerts(&self) -> Vec<ExpiryAlert> {
        self.get_expiry_alerts_internal()
    }

    pub fn recall(&self, req: RecallRequest) -> Result<RecallRecord> {
        let batch_no = req.batch_no.trim().to_string();
        if batch_no.is_empty() {
            return Err(DrugTraceError::EmptyBatchNo);
        }

        let batch_id = {
            let batch_no_map = self.batch_no_to_id.lock().unwrap();
            batch_no_map.get(&batch_no).cloned()
        };

        let batch_id = match batch_id {
            Some(id) => id,
            None => return Err(DrugTraceError::BatchNotFound(batch_no)),
        };

        let (inbound_record, batch) = {
            let batches = self.batches.lock().unwrap();
            let inbound_records = self.inbound_records.lock().unwrap();

            let batch = batches.get(&batch_id).cloned();
            let inbound = inbound_records
                .iter()
                .find(|r| r.batch_id == batch_id)
                .cloned();

            (inbound, batch)
        };

        if batch.is_none() {
            return Err(DrugTraceError::BatchNotFound(batch_no));
        }

        let batch = batch.unwrap();
        let drug_name = batch.drug_name.clone();
        let supplier = batch.supplier.clone();

        let in_stock_batches = {
            let mut batches = self.batches.lock().unwrap();
            let mut result = Vec::new();

            for b in batches.values_mut() {
                if b.batch_no == batch_no && !b.recall_flag {
                    b.recall_flag = true;
                    if b.remaining_quantity > 0 {
                        b.status = BatchStatus::Frozen;
                    } else {
                        b.status = BatchStatus::Recalled;
                    }
                    result.push(b.clone());
                }
            }
            result
        };

        let outbound_records = {
            let outbound = self.outbound_records.lock().unwrap();
            outbound
                .iter()
                .filter(|r| r.details.iter().any(|d| d.batch_no == batch_no))
                .cloned()
                .collect()
        };

        Ok(RecallRecord {
            id: Uuid::new_v4(),
            batch_no,
            drug_name,
            supplier,
            inbound: inbound_record,
            in_stock_batches,
            outbound_records,
            created_at: Utc::now(),
        })
    }

    pub fn get_all_batches(&self) -> Vec<DrugBatch> {
        let batches = self.batches.lock().unwrap();
        batches.values().cloned().collect()
    }

    pub fn get_batch(&self, batch_no: &str) -> Option<DrugBatch> {
        let batch_no_map = self.batch_no_to_id.lock().unwrap();
        let batch_id = batch_no_map.get(batch_no)?;

        let batches = self.batches.lock().unwrap();
        batches.get(batch_id).cloned()
    }

    pub fn get_inbound_records(&self) -> Vec<InboundRecord> {
        let records = self.inbound_records.lock().unwrap();
        records.clone()
    }

    pub fn get_outbound_records(&self) -> Vec<OutboundRecord> {
        let records = self.outbound_records.lock().unwrap();
        records.clone()
    }
}
