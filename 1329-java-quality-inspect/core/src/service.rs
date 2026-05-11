use chrono::Utc;
use std::error::Error;
use uuid::Uuid;

use crate::aql::{AqlConfig, AqlLevel};
use crate::models::{
    CreateOrderRequest, DefectRecord, InspectItemRequest, InspectionOrder, InspectionResult,
    InspectionStatus, RecordDefectRequest, ReinspectionRequest, UpdateDefectRequest,
};
use crate::storage::InMemoryStorage;

#[derive(Clone)]
pub struct QaService {
    storage: InMemoryStorage,
    aql_config: AqlConfig,
}

impl QaService {
    pub fn new() -> Self {
        QaService {
            storage: InMemoryStorage::new(),
            aql_config: AqlConfig::new(),
        }
    }

    pub fn create_order(&self, request: CreateOrderRequest) -> Result<InspectionOrder, Box<dyn Error>> {
        let aql_level = AqlLevel::from_str(&request.aql_level)
            .ok_or_else(|| format!("Invalid AQL level: {}", request.aql_level))?;

        let sample_size = self.aql_config.get_sample_size(request.batch_size);
        let ac = self.aql_config.get_ac(sample_size, aql_level);

        let now = Utc::now();
        let order = InspectionOrder {
            id: Uuid::new_v4(),
            product_name: request.product_name,
            batch_size: request.batch_size,
            sample_size,
            aql_level,
            ac,
            status: InspectionStatus::Pending,
            inspector: request.inspector,
            created_at: now,
            updated_at: now,
            is_reinspection: false,
            parent_order_id: None,
            defect_count: 0,
            inspection_count: 0,
        };

        self.storage.save_order(order.clone());
        Ok(order)
    }

    pub fn start_inspection(&self, order_id: Uuid) -> Result<InspectionOrder, Box<dyn Error>> {
        let mut order = self
            .storage
            .get_order(order_id)
            .ok_or_else(|| format!("Order not found: {}", order_id))?;

        match order.status {
            InspectionStatus::Pending => {
                order.status = InspectionStatus::Inspecting;
            }
            InspectionStatus::ReinspectionRequested => {
                order.status = InspectionStatus::Reinspecting;
            }
            _ => return Err(format!("Invalid status for starting inspection: {:?}", order.status).into()),
        }

        order.updated_at = Utc::now();
        self.storage.save_order(order.clone());
        Ok(order)
    }

    pub fn record_defect(&self, request: RecordDefectRequest) -> Result<DefectRecord, Box<dyn Error>> {
        let mut order = self
            .storage
            .get_order(request.order_id)
            .ok_or_else(|| format!("Order not found: {}", request.order_id))?;

        match order.status {
            InspectionStatus::Inspecting | InspectionStatus::Reinspecting => {}
            _ => return Err(format!("Cannot record defect in status: {:?}", order.status).into()),
        }

        if request.sample_item_number > order.sample_size {
            return Err(format!(
                "Sample item number {} exceeds sample size {}",
                request.sample_item_number, order.sample_size
            )
            .into());
        }

        let now = Utc::now();
        let defect = DefectRecord {
            id: Uuid::new_v4(),
            order_id: request.order_id,
            defect_type: request.defect_type,
            description: request.description,
            sample_item_number: request.sample_item_number,
            recorded_at: now,
            updated_at: now,
        };

        self.storage.save_defect(defect.clone());

        order.defect_count += 1;
        order.updated_at = now;
        self.storage.save_order(order);

        Ok(defect)
    }

    pub fn inspect_item(&self, request: InspectItemRequest) -> Result<Option<DefectRecord>, Box<dyn Error>> {
        let mut order = self
            .storage
            .get_order(request.order_id)
            .ok_or_else(|| format!("Order not found: {}", request.order_id))?;

        match order.status {
            InspectionStatus::Inspecting | InspectionStatus::Reinspecting => {}
            _ => return Err(format!("Cannot inspect item in status: {:?}", order.status).into()),
        }

        if request.item_number > order.sample_size {
            return Err(format!(
                "Sample item number {} exceeds sample size {}",
                request.item_number, order.sample_size
            )
            .into());
        }

        let _existing_inspections = self.storage.get_defects_for_order(order.id).len() as u32;
        if request.item_number <= order.inspection_count {
            return Err(format!(
                "Item {} already inspected. Current inspection count: {}",
                request.item_number, order.inspection_count
            )
            .into());
        }

        let now = Utc::now();
        let defect = if request.is_defective {
            let defect_type = request
                .defect_type
                .ok_or_else(|| "Defect type is required for defective items".to_string())?;
            let description = request
                .defect_description
                .ok_or_else(|| "Defect description is required for defective items".to_string())?;

            let defect = DefectRecord {
                id: Uuid::new_v4(),
                order_id: request.order_id,
                defect_type,
                description,
                sample_item_number: request.item_number,
                recorded_at: now,
                updated_at: now,
            };

            self.storage.save_defect(defect.clone());
            Some(defect)
        } else {
            None
        };

        order.inspection_count += 1;
        if defect.is_some() {
            order.defect_count += 1;
        }
        order.updated_at = now;
        self.storage.save_order(order);

        Ok(defect)
    }

    pub fn complete_inspection(&self, order_id: Uuid) -> Result<InspectionResult, Box<dyn Error>> {
        let mut order = self
            .storage
            .get_order(order_id)
            .ok_or_else(|| format!("Order not found: {}", order_id))?;

        match order.status {
            InspectionStatus::Inspecting | InspectionStatus::Reinspecting => {}
            _ => return Err(format!("Cannot complete inspection in status: {:?}", order.status).into()),
        }

        if order.inspection_count < order.sample_size {
            return Err(format!(
                "Not all items inspected. Inspected: {}, Required: {}",
                order.inspection_count, order.sample_size
            )
            .into());
        }

        let defect_count = order.defect_count;
        let ac = order.ac;
        let sample_size = order.sample_size;
        let accepted = defect_count <= ac;
        let is_final = order.is_reinspection;

        order.status = if is_final {
            if accepted {
                InspectionStatus::FinalAccepted
            } else {
                InspectionStatus::FinalRejected
            }
        } else if accepted {
            InspectionStatus::Completed
        } else {
            InspectionStatus::Rejected
        };

        order.updated_at = Utc::now();
        self.storage.save_order(order);

        Ok(InspectionResult {
            accepted,
            defect_count,
            ac,
            sample_size,
            is_final,
        })
    }

    pub fn request_reinspection(&self, request: ReinspectionRequest) -> Result<InspectionOrder, Box<dyn Error>> {
        let parent_order = self
            .storage
            .get_order(request.order_id)
            .ok_or_else(|| format!("Order not found: {}", request.order_id))?;

        if parent_order.status != InspectionStatus::Rejected {
            return Err(format!(
                "Reinspection can only be requested for rejected orders. Current status: {:?}",
                parent_order.status
            )
            .into());
        }

        let new_sample_size = parent_order.sample_size * 2;
        let ac = parent_order.ac;

        let now = Utc::now();
        let reinspection_order = InspectionOrder {
            id: Uuid::new_v4(),
            product_name: parent_order.product_name.clone(),
            batch_size: parent_order.batch_size,
            sample_size: new_sample_size,
            aql_level: parent_order.aql_level,
            ac,
            status: InspectionStatus::ReinspectionRequested,
            inspector: request.inspector,
            created_at: now,
            updated_at: now,
            is_reinspection: true,
            parent_order_id: Some(parent_order.id),
            defect_count: 0,
            inspection_count: 0,
        };

        self.storage.save_order(reinspection_order.clone());
        Ok(reinspection_order)
    }

    pub fn update_defect_description(&self, request: UpdateDefectRequest) -> Result<DefectRecord, Box<dyn Error>> {
        let mut defect = self
            .storage
            .get_defect(request.defect_id)
            .ok_or_else(|| format!("Defect not found: {}", request.defect_id))?;

        defect.description = request.new_description;
        defect.updated_at = Utc::now();

        self.storage.save_defect(defect.clone());
        Ok(defect)
    }

    pub fn get_order(&self, order_id: Uuid) -> Option<InspectionOrder> {
        self.storage.get_order(order_id)
    }

    pub fn get_all_orders(&self) -> Vec<InspectionOrder> {
        self.storage.get_all_orders()
    }

    pub fn get_defects_for_order(&self, order_id: Uuid) -> Vec<DefectRecord> {
        self.storage.get_defects_for_order(order_id)
    }

    pub fn get_defect(&self, defect_id: Uuid) -> Option<DefectRecord> {
        self.storage.get_defect(defect_id)
    }
}

impl Default for QaService {
    fn default() -> Self {
        Self::new()
    }
}
