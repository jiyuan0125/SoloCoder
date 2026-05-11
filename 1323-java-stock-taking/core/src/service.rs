use std::collections::HashMap;
use std::sync::Arc;

use chrono::{DateTime, Datelike, Utc};
use rust_decimal::Decimal;
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::errors::{StockError, StockResult};
use crate::models::{
    ApproveDifferenceRequest, BatchState, CountRecord, CreateCountRecordRequest, CreateBatchRequest,
    DifferenceState, DifferenceType, Material, StockBatch, StockDifference, SubmitDifferenceRequest,
};

pub struct InventoryService {
    materials: RwLock<HashMap<String, Material>>,
    batches: RwLock<HashMap<String, StockBatch>>,
    count_records: RwLock<HashMap<String, CountRecord>>,
    differences: RwLock<HashMap<Uuid, StockDifference>>,
    config: ServiceConfig,
}

#[derive(Debug, Clone)]
pub struct ServiceConfig {
    pub auto_approve_threshold: Decimal,
}

impl Default for ServiceConfig {
    fn default() -> Self {
        Self {
            auto_approve_threshold: Decimal::new(100, 0),
        }
    }
}

impl InventoryService {
    pub fn new(config: ServiceConfig) -> Arc<Self> {
        let service = Arc::new(Self {
            materials: RwLock::new(HashMap::new()),
            batches: RwLock::new(HashMap::new()),
            count_records: RwLock::new(HashMap::new()),
            differences: RwLock::new(HashMap::new()),
            config,
        });
        service
    }

    pub fn new_with_default_config() -> Arc<Self> {
        Self::new(ServiceConfig::default())
    }

    pub fn config(&self) -> &ServiceConfig {
        &self.config
    }

    pub async fn add_material(&self, material: Material) {
        let mut materials = self.materials.write().await;
        materials.insert(material.code.clone(), material);
    }

    pub async fn get_material(&self, code: &str) -> Option<Material> {
        let materials = self.materials.read().await;
        materials.get(code).cloned()
    }

    pub async fn list_materials(&self) -> Vec<Material> {
        let materials = self.materials.read().await;
        materials.values().cloned().collect()
    }

    pub async fn create_batch(&self, _req: CreateBatchRequest) -> StockResult<StockBatch> {
        let batch_no = generate_batch_no();
        let batch = StockBatch {
            batch_no: batch_no.clone(),
            state: BatchState::InProgress,
            created_at: Utc::now(),
            completed_at: None,
        };

        let mut batches = self.batches.write().await;
        batches.insert(batch_no, batch.clone());
        Ok(batch)
    }

    pub async fn get_batch(&self, batch_no: &str) -> Option<StockBatch> {
        let batches = self.batches.read().await;
        batches.get(batch_no).cloned()
    }

    pub async fn list_batches(&self) -> Vec<StockBatch> {
        let batches = self.batches.read().await;
        batches.values().cloned().collect()
    }

    pub async fn cancel_batch(&self, batch_no: &str) -> StockResult<StockBatch> {
        let mut batches = self.batches.write().await;
        let batch = batches.get_mut(batch_no).ok_or_else(|| {
            StockError::BatchNotFound(batch_no.to_string())
        })?;

        if batch.state != BatchState::InProgress {
            return Err(StockError::InvalidBatchState(
                "只有进行中的批次可以取消".to_string(),
            ));
        }

        batch.state = BatchState::Cancelled;
        Ok(batch.clone())
    }

    pub async fn complete_batch(&self, batch_no: &str) -> StockResult<StockBatch> {
        let mut batches = self.batches.write().await;
        let batch = batches.get_mut(batch_no).ok_or_else(|| {
            StockError::BatchNotFound(batch_no.to_string())
        })?;

        if batch.state != BatchState::InProgress {
            return Err(StockError::InvalidBatchState(
                "只有进行中的批次可以完成".to_string(),
            ));
        }

        let differences = self.differences.read().await;
        let batch_diffs: Vec<_> = differences.values()
            .filter(|d| d.batch_no == batch_no)
            .collect();

        for diff in &batch_diffs {
            if diff.state != DifferenceState::Adjusted && diff.state != DifferenceState::AutoApproved {
                return Err(StockError::InvalidBatchState(
                    "存在未处理的差异，请先处理所有差异后再完成批次".to_string(),
                ));
            }
        }

        batch.state = BatchState::Completed;
        batch.completed_at = Some(Utc::now());
        Ok(batch.clone())
    }

    pub async fn submit_count(&self, req: CreateCountRecordRequest) -> StockResult<StockDifference> {
        {
            let batches = self.batches.read().await;
            let batch = batches.get(&req.batch_no).ok_or_else(|| {
                StockError::BatchNotFound(req.batch_no.clone())
            })?;

            if batch.state != BatchState::InProgress {
                return Err(StockError::InvalidBatchState(
                    "只能在进行中的批次录入盘点数据".to_string(),
                ));
            }
        }

        let material = {
            let materials = self.materials.read().await;
            materials.get(&req.material_code).cloned().ok_or_else(|| {
                StockError::MaterialNotFound(req.material_code.clone())
            })?
        };

        let record = CountRecord {
            id: Uuid::new_v4(),
            batch_no: req.batch_no.clone(),
            material_code: req.material_code.clone(),
            actual_quantity: req.actual_quantity,
            operator: req.operator.clone(),
            counted_at: Utc::now(),
        };

        {
            let mut records = self.count_records.write().await;
            records.insert(record.id.to_string(), record);
        }

        let diff_qty = req.actual_quantity - material.book_quantity;
        let diff_amt = Decimal::from(diff_qty) * material.price;
        let diff_type = if diff_qty >= 0 {
            DifferenceType::Over
        } else {
            DifferenceType::Shortage
        };

        let diff = StockDifference {
            id: Uuid::new_v4(),
            batch_no: req.batch_no.clone(),
            material_code: req.material_code.clone(),
            book_quantity: material.book_quantity,
            actual_quantity: req.actual_quantity,
            difference_quantity: diff_qty,
            difference_amount: diff_amt,
            difference_type: diff_type,
            state: DifferenceState::PendingAutoReview,
            reason: None,
            operator: req.operator.clone(),
            approver: None,
            created_at: Utc::now(),
            approved_at: None,
        };

        let should_auto_approve = diff.difference_amount.abs() <= self.config.auto_approve_threshold;
        let mut diff = diff;

        if should_auto_approve {
            diff.state = DifferenceState::AutoApproved;
            self.adjust_inventory(&diff).await?;
            diff.state = DifferenceState::Adjusted;
        } else {
            diff.state = DifferenceState::PendingApproval;
        }

        {
            let mut differences = self.differences.write().await;
            differences.insert(diff.id, diff.clone());
        }

        Ok(diff)
    }

    pub async fn list_batch_records(&self, batch_no: &str) -> Vec<CountRecord> {
        let records = self.count_records.read().await;
        records.values()
            .filter(|r| r.batch_no == batch_no)
            .cloned()
            .collect()
    }

    pub async fn list_batch_differences(&self, batch_no: &str) -> Vec<StockDifference> {
        let differences = self.differences.read().await;
        differences.values()
            .filter(|d| d.batch_no == batch_no)
            .cloned()
            .collect()
    }

    pub async fn get_difference(&self, id: &Uuid) -> Option<StockDifference> {
        let differences = self.differences.read().await;
        differences.get(id).cloned()
    }

    pub async fn submit_difference(
        &self,
        id: &Uuid,
        req: SubmitDifferenceRequest,
    ) -> StockResult<StockDifference> {
        if req.reason.trim().is_empty() {
            return Err(StockError::ReasonRequired);
        }

        let mut differences = self.differences.write().await;
        let diff = differences.get_mut(id).ok_or_else(|| {
            StockError::DifferenceNotFound(id.to_string())
        })?;

        if diff.state != DifferenceState::PendingApproval && diff.state != DifferenceState::Rejected {
            return Err(StockError::InvalidDifferenceState(
                "只有待审批或已驳回的差异可以提交".to_string(),
            ));
        }

        {
            let batches = self.batches.read().await;
            if let Some(batch) = batches.get(&diff.batch_no) {
                if batch.state != BatchState::InProgress {
                    return Err(StockError::InvalidBatchState(
                        "批次已完成或已取消，无法操作差异".to_string(),
                    ));
                }
            }
        }

        diff.reason = Some(req.reason);
        diff.state = DifferenceState::Submitted;
        Ok(diff.clone())
    }

    pub async fn approve_difference(
        &self,
        id: &Uuid,
        req: ApproveDifferenceRequest,
    ) -> StockResult<StockDifference> {
        let mut differences = self.differences.write().await;
        let diff = differences.get_mut(id).ok_or_else(|| {
            StockError::DifferenceNotFound(id.to_string())
        })?;

        if diff.state != DifferenceState::Submitted {
            return Err(StockError::InvalidDifferenceState(
                "只有已提交的差异可以审批".to_string(),
            ));
        }

        {
            let batches = self.batches.read().await;
            if let Some(batch) = batches.get(&diff.batch_no) {
                if batch.state != BatchState::InProgress {
                    return Err(StockError::InvalidBatchState(
                        "批次已完成或已取消，无法操作差异".to_string(),
                    ));
                }
            }
        }

        diff.approver = Some(req.approver);
        diff.approved_at = Some(Utc::now());

        if req.approved {
            diff.state = DifferenceState::Approved;
            let diff_clone = diff.clone();
            drop(differences);
            self.adjust_inventory(&diff_clone).await?;
            let mut differences = self.differences.write().await;
            let diff = differences.get_mut(id).ok_or_else(|| {
                StockError::DifferenceNotFound(id.to_string())
            })?;
            diff.state = DifferenceState::Adjusted;
            Ok(diff.clone())
        } else {
            diff.state = DifferenceState::Rejected;
            Ok(diff.clone())
        }
    }

    async fn adjust_inventory(&self, diff: &StockDifference) -> StockResult<()> {
        let mut materials = self.materials.write().await;
        let material = materials.get_mut(&diff.material_code).ok_or_else(|| {
            StockError::MaterialNotFound(diff.material_code.clone())
        })?;
        material.book_quantity = diff.actual_quantity;
        Ok(())
    }
}

fn generate_batch_no() -> String {
    let now: DateTime<Utc> = Utc::now();
    format!("PD{}{:04}{:02}", now.year(), now.month(), now.day())
}
