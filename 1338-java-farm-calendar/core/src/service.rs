use crate::error::{AppError, AppResult};
use crate::models::*;
use crate::date_utils::*;
use crate::advice::get_advice_for_stage;
use chrono::{Utc, NaiveDate};
use std::collections::HashMap;
use std::sync::RwLock;
use uuid::Uuid;

pub trait DataStore: Send + Sync {
    fn get_crop(&self, id: Uuid) -> AppResult<Crop>;
    fn get_all_crops(&self) -> Vec<Crop>;
    fn add_crop(&self, crop: Crop) -> AppResult<()>;

    fn get_region(&self, id: Uuid) -> AppResult<Region>;
    fn get_all_regions(&self) -> Vec<Region>;
    fn add_region(&self, region: Region) -> AppResult<()>;

    fn get_plot(&self, id: Uuid) -> AppResult<Plot>;
    fn get_all_plots(&self) -> Vec<Plot>;
    fn add_plot(&self, plot: Plot) -> AppResult<()>;

    fn get_sowing_plan(&self, id: Uuid) -> AppResult<SowingPlan>;
    fn get_all_sowing_plans(&self) -> Vec<SowingPlan>;
    fn get_sowing_plans_by_plot(&self, plot_id: Uuid) -> Vec<SowingPlan>;
    fn add_sowing_plan(&self, plan: SowingPlan) -> AppResult<()>;
}

pub struct InMemoryStore {
    crops: RwLock<HashMap<Uuid, Crop>>,
    regions: RwLock<HashMap<Uuid, Region>>,
    plots: RwLock<HashMap<Uuid, Plot>>,
    sowing_plans: RwLock<HashMap<Uuid, SowingPlan>>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self {
            crops: RwLock::new(HashMap::new()),
            regions: RwLock::new(HashMap::new()),
            plots: RwLock::new(HashMap::new()),
            sowing_plans: RwLock::new(HashMap::new()),
        }
    }
}

impl Default for InMemoryStore {
    fn default() -> Self {
        Self::new()
    }
}

impl DataStore for InMemoryStore {
    fn get_crop(&self, id: Uuid) -> AppResult<Crop> {
        self.crops.read().unwrap()
            .get(&id)
            .cloned()
            .ok_or_else(|| AppError::NotFound(format!("Crop with id {} not found", id)))
    }

    fn get_all_crops(&self) -> Vec<Crop> {
        self.crops.read().unwrap().values().cloned().collect()
    }

    fn add_crop(&self, crop: Crop) -> AppResult<()> {
        self.crops.write().unwrap().insert(crop.id, crop);
        Ok(())
    }

    fn get_region(&self, id: Uuid) -> AppResult<Region> {
        self.regions.read().unwrap()
            .get(&id)
            .cloned()
            .ok_or_else(|| AppError::NotFound(format!("Region with id {} not found", id)))
    }

    fn get_all_regions(&self) -> Vec<Region> {
        self.regions.read().unwrap().values().cloned().collect()
    }

    fn add_region(&self, region: Region) -> AppResult<()> {
        self.regions.write().unwrap().insert(region.id, region);
        Ok(())
    }

    fn get_plot(&self, id: Uuid) -> AppResult<Plot> {
        self.plots.read().unwrap()
            .get(&id)
            .cloned()
            .ok_or_else(|| AppError::NotFound(format!("Plot with id {} not found", id)))
    }

    fn get_all_plots(&self) -> Vec<Plot> {
        self.plots.read().unwrap().values().cloned().collect()
    }

    fn add_plot(&self, plot: Plot) -> AppResult<()> {
        self.plots.write().unwrap().insert(plot.id, plot);
        Ok(())
    }

    fn get_sowing_plan(&self, id: Uuid) -> AppResult<SowingPlan> {
        self.sowing_plans.read().unwrap()
            .get(&id)
            .cloned()
            .ok_or_else(|| AppError::NotFound(format!("Sowing plan with id {} not found", id)))
    }

    fn get_all_sowing_plans(&self) -> Vec<SowingPlan> {
        self.sowing_plans.read().unwrap().values().cloned().collect()
    }

    fn get_sowing_plans_by_plot(&self, plot_id: Uuid) -> Vec<SowingPlan> {
        self.sowing_plans.read().unwrap()
            .values()
            .filter(|p| p.plot_id == plot_id)
            .cloned()
            .collect()
    }

    fn add_sowing_plan(&self, plan: SowingPlan) -> AppResult<()> {
        self.sowing_plans.write().unwrap().insert(plan.id, plan);
        Ok(())
    }
}

pub struct AgriculturalService<T: DataStore> {
    store: T,
}

impl<T: DataStore> AgriculturalService<T> {
    pub fn new(store: T) -> Self {
        Self { store }
    }

    pub fn create_crop(&self, request: CreateCropRequest) -> AppResult<Crop> {
        let stages: Vec<CropStage> = request.stages.into_iter()
            .map(|s| CropStage {
                name: s.name,
                days: s.days,
                description: s.description,
            })
            .collect();

        let crop = Crop::new(
            request.name,
            request.total_growth_days,
            stages,
            request.description,
        );

        if !crop.validate_stage_days() {
            let actual_days: u32 = crop.stages.iter().map(|s| s.days).sum();
            return Err(AppError::Validation(format!(
                "Stage days sum ({}) does not equal total growth days ({})",
                actual_days, crop.total_growth_days
            )));
        }

        self.store.add_crop(crop.clone())?;
        Ok(crop)
    }

    pub fn get_crop(&self, id: Uuid) -> AppResult<Crop> {
        self.store.get_crop(id)
    }

    pub fn get_all_crops(&self) -> Vec<Crop> {
        self.store.get_all_crops()
    }

    pub fn create_region(&self, request: CreateRegionRequest) -> AppResult<Region> {
        for &month in &request.suitable_sowing_months {
            if month < 1 || month > 12 {
                return Err(AppError::Validation(format!("Invalid month: {}", month)));
            }
        }

        let region = Region::new(
            request.name,
            request.average_effective_temperature,
            request.suitable_sowing_months,
            request.description,
        );

        self.store.add_region(region.clone())?;
        Ok(region)
    }

    pub fn get_region(&self, id: Uuid) -> AppResult<Region> {
        self.store.get_region(id)
    }

    pub fn get_all_regions(&self) -> Vec<Region> {
        self.store.get_all_regions()
    }

    pub fn create_plot(&self, request: CreatePlotRequest) -> AppResult<Plot> {
        self.store.get_region(request.region_id)?;

        let plot = Plot::new(
            request.name,
            request.region_id,
            request.area,
            request.unit,
            request.description,
        );

        self.store.add_plot(plot.clone())?;
        Ok(plot)
    }

    pub fn get_plot(&self, id: Uuid) -> AppResult<Plot> {
        self.store.get_plot(id)
    }

    pub fn get_all_plots(&self) -> Vec<Plot> {
        self.store.get_all_plots()
    }

    pub fn create_sowing_plan(&self, request: CreateSowingPlanRequest) -> AppResult<SowingPlan> {
        let crop = self.store.get_crop(request.crop_id)?;
        let plot = self.store.get_plot(request.plot_id)?;
        let region = self.store.get_region(plot.region_id)?;
        let sowing_date = parse_date(&request.sowing_date)?;

        let warnings = self.check_sowing_warnings(&region, sowing_date);
        let stage_schedules = self.calculate_stage_schedules(&crop, sowing_date);
        let (start_date, end_date) = self.calculate_growth_period(&crop, sowing_date);

        let existing_plans = self.store.get_sowing_plans_by_plot(request.plot_id);
        for existing in &existing_plans {
            if date_overlaps(start_date, end_date, existing.start_date, existing.end_date) {
                return Err(AppError::Conflict(format!(
                    "Sowing plan conflicts with existing plan {} on plot {} ({} to {})",
                    existing.id, plot.name, existing.start_date, existing.end_date
                )));
            }
        }

        let plan = SowingPlan {
            id: Uuid::new_v4(),
            crop_id: request.crop_id,
            plot_id: request.plot_id,
            sowing_date,
            start_date,
            end_date,
            stage_schedules,
            warnings,
            created_at: Utc::now(),
        };

        self.store.add_sowing_plan(plan.clone())?;
        Ok(plan)
    }

    pub fn batch_create_sowing_plans(&self, request: BatchCreateSowingPlanRequest) -> AppResult<BatchCreateSowingPlanResult> {
        let mut results: HashMap<Uuid, SowingPlanResult> = HashMap::new();
        let mut has_conflicts = false;
        let temp_id_mapping: HashMap<Uuid, CreateSowingPlanRequest> = request.plans.iter()
            .map(|r| (Uuid::new_v4(), r.clone()))
            .collect();

        let validated_plans: Vec<(Uuid, SowingPlan, CreateSowingPlanRequest)> = temp_id_mapping.iter()
            .filter_map(|(temp_id, req)| {
                match self.validate_sowing_plan(req) {
                    Ok(plan) => Some((*temp_id, plan, req.clone())),
                    Err(e) => {
                        results.insert(*temp_id, SowingPlanResult {
                            plan: None,
                            errors: vec![e.to_string()],
                        });
                        None
                    }
                }
            })
            .collect();

        for (i, (temp_id_i, plan_i, req_i)) in validated_plans.iter().enumerate() {
            let mut has_new_conflict = false;

            for (j, (_, plan_j, _)) in validated_plans.iter().enumerate() {
                if i != j && plan_i.plot_id == plan_j.plot_id {
                    if date_overlaps(plan_i.start_date, plan_i.end_date, plan_j.start_date, plan_j.end_date) {
                        let error = format!(
                            "Conflict with another batch plan: plot {} has overlapping period ({} to {})",
                            self.get_plot_name(plan_i.plot_id),
                            plan_j.start_date,
                            plan_j.end_date
                        );

                        if let Some(result) = results.get_mut(temp_id_i) {
                            result.errors.push(error.clone());
                        } else {
                            results.insert(*temp_id_i, SowingPlanResult {
                                plan: None,
                                errors: vec![error],
                            });
                        }
                        has_conflicts = true;
                        has_new_conflict = true;
                    }
                }
            }

            if !has_new_conflict {
                match self.create_sowing_plan(req_i.clone()) {
                    Ok(created_plan) => {
                        results.insert(*temp_id_i, SowingPlanResult {
                            plan: Some(created_plan),
                            errors: vec![],
                        });
                    }
                    Err(e) => {
                        if let AppError::Conflict(_) = &e {
                            has_conflicts = true;
                        }
                        results.insert(*temp_id_i, SowingPlanResult {
                            plan: None,
                            errors: vec![e.to_string()],
                        });
                    }
                }
            }
        }

        Ok(BatchCreateSowingPlanResult {
            results,
            has_conflicts,
        })
    }

    fn validate_sowing_plan(&self, request: &CreateSowingPlanRequest) -> AppResult<SowingPlan> {
        let crop = self.store.get_crop(request.crop_id)?;
        let plot = self.store.get_plot(request.plot_id)?;
        let region = self.store.get_region(plot.region_id)?;
        let sowing_date = parse_date(&request.sowing_date)?;

        let warnings = self.check_sowing_warnings(&region, sowing_date);
        let stage_schedules = self.calculate_stage_schedules(&crop, sowing_date);
        let (start_date, end_date) = self.calculate_growth_period(&crop, sowing_date);

        Ok(SowingPlan {
            id: Uuid::new_v4(),
            crop_id: request.crop_id,
            plot_id: request.plot_id,
            sowing_date,
            start_date,
            end_date,
            stage_schedules,
            warnings,
            created_at: Utc::now(),
        })
    }

    fn get_plot_name(&self, plot_id: Uuid) -> String {
        self.store.get_plot(plot_id)
            .map(|p| p.name)
            .unwrap_or_else(|_| "Unknown".to_string())
    }

    pub fn get_sowing_plan(&self, id: Uuid) -> AppResult<SowingPlan> {
        self.store.get_sowing_plan(id)
    }

    pub fn get_all_sowing_plans(&self) -> Vec<SowingPlan> {
        self.store.get_all_sowing_plans()
    }

    pub fn get_sowing_plans_by_plot(&self, plot_id: Uuid) -> Vec<SowingPlan> {
        self.store.get_sowing_plans_by_plot(plot_id)
    }

    fn check_sowing_warnings(&self, region: &Region, sowing_date: NaiveDate) -> Vec<String> {
        let mut warnings = Vec::new();
        let sowing_month = get_month(sowing_date);

        if !region.is_suitable_month(sowing_month) {
            let suitable_months: Vec<String> = region.suitable_sowing_months
                .iter()
                .map(|&m| format_month(m))
                .collect();
            warnings.push(format!(
                "Sowing date is not in the suitable months for region {}. Suitable months: {}",
                region.name,
                suitable_months.join(", ")
            ));
        }

        warnings
    }

    fn calculate_stage_schedules(&self, crop: &Crop, sowing_date: NaiveDate) -> Vec<StageSchedule> {
        let mut schedules = Vec::new();
        let mut current_date = sowing_date;

        for stage in &crop.stages {
            let start_date = current_date;
            let end_date = add_days(start_date, stage.days.saturating_sub(1));

            schedules.push(StageSchedule {
                stage_name: stage.name.clone(),
                start_date,
                end_date,
                days: stage.days,
                advice: get_advice_for_stage(&stage.name),
            });

            current_date = add_days(end_date, 1);
        }

        schedules
    }

    fn calculate_growth_period(&self, crop: &Crop, sowing_date: NaiveDate) -> (NaiveDate, NaiveDate) {
        let start_date = sowing_date;
        let end_date = add_days(start_date, crop.total_growth_days.saturating_sub(1));
        (start_date, end_date)
    }
}
