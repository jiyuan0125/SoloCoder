use chrono::{DateTime, Utc, NaiveDate};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CropStage {
    pub name: String,
    pub days: u32,
    pub description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Crop {
    pub id: Uuid,
    pub name: String,
    pub total_growth_days: u32,
    pub stages: Vec<CropStage>,
    pub description: String,
}

impl Crop {
    pub fn new(name: String, total_growth_days: u32, stages: Vec<CropStage>, description: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            total_growth_days,
            stages,
            description,
        }
    }

    pub fn validate_stage_days(&self) -> bool {
        let total: u32 = self.stages.iter().map(|s| s.days).sum();
        total == self.total_growth_days
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Region {
    pub id: Uuid,
    pub name: String,
    pub average_effective_temperature: f64,
    pub suitable_sowing_months: Vec<u32>,
    pub description: String,
}

impl Region {
    pub fn new(name: String, average_effective_temperature: f64, suitable_sowing_months: Vec<u32>, description: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            average_effective_temperature,
            suitable_sowing_months,
            description,
        }
    }

    pub fn is_suitable_month(&self, month: u32) -> bool {
        self.suitable_sowing_months.contains(&month)
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Plot {
    pub id: Uuid,
    pub name: String,
    pub region_id: Uuid,
    pub area: f64,
    pub unit: String,
    pub description: String,
}

impl Plot {
    pub fn new(name: String, region_id: Uuid, area: f64, unit: String, description: String) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            region_id,
            area,
            unit,
            description,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StageSchedule {
    pub stage_name: String,
    pub start_date: NaiveDate,
    pub end_date: NaiveDate,
    pub days: u32,
    pub advice: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SowingPlan {
    pub id: Uuid,
    pub crop_id: Uuid,
    pub plot_id: Uuid,
    pub sowing_date: NaiveDate,
    pub start_date: NaiveDate,
    pub end_date: NaiveDate,
    pub stage_schedules: Vec<StageSchedule>,
    pub warnings: Vec<String>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateCropRequest {
    pub name: String,
    pub total_growth_days: u32,
    pub stages: Vec<CreateCropStageRequest>,
    pub description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateCropStageRequest {
    pub name: String,
    pub days: u32,
    pub description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateRegionRequest {
    pub name: String,
    pub average_effective_temperature: f64,
    pub suitable_sowing_months: Vec<u32>,
    pub description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreatePlotRequest {
    pub name: String,
    pub region_id: Uuid,
    pub area: f64,
    pub unit: String,
    pub description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateSowingPlanRequest {
    pub crop_id: Uuid,
    pub plot_id: Uuid,
    pub sowing_date: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BatchCreateSowingPlanRequest {
    pub plans: Vec<CreateSowingPlanRequest>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SowingPlanResult {
    pub plan: Option<SowingPlan>,
    pub errors: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BatchCreateSowingPlanResult {
    pub results: HashMap<Uuid, SowingPlanResult>,
    pub has_conflicts: bool,
}
