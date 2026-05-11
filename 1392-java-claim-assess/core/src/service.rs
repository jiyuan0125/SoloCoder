use crate::{
    AppError, AppResult, Claim, ClaimStatus, InMemoryStore, RepairItem, RepairItemType,
    TOTAL_LOSS_THRESHOLD_PERCENT, Vehicle,
};
use rust_decimal::Decimal;

#[derive(Debug, Clone)]
pub struct ClaimService {
    store: InMemoryStore,
}

impl ClaimService {
    pub fn new(store: InMemoryStore) -> Self {
        Self { store }
    }

    pub fn create_vehicle(
        &self,
        plate_number: String,
        model: String,
        actual_value: Decimal,
    ) -> Vehicle {
        let vehicle = Vehicle::new(plate_number, model, actual_value);
        self.store.save_vehicle(vehicle.clone());
        vehicle
    }

    pub fn get_vehicle(&self, id: &str) -> AppResult<Vehicle> {
        self.store
            .get_vehicle(id)
            .ok_or_else(|| AppError::VehicleNotFound(id.to_string()))
    }

    pub fn get_vehicle_by_plate(&self, plate_number: &str) -> AppResult<Vehicle> {
        self.store
            .get_vehicle_by_plate(plate_number)
            .ok_or_else(|| AppError::VehicleNotFound(plate_number.to_string()))
    }

    pub fn list_vehicles(&self) -> Vec<Vehicle> {
        self.store.list_vehicles()
    }

    pub fn update_vehicle_actual_value(&self, vehicle_id: &str, actual_value: Decimal) -> AppResult<Vehicle> {
        let vehicle = self.get_vehicle(vehicle_id)?;

        let claims = self.store.list_claims_by_vehicle(vehicle_id);
        let has_unsettled = claims
            .iter()
            .any(|c| c.status != ClaimStatus::Settled);

        if has_unsettled {
            return Err(AppError::HasUnsettledClaim);
        }

        let mut updated = vehicle.clone();
        updated.actual_value = actual_value;
        self.store.save_vehicle(updated.clone());
        Ok(updated)
    }

    pub fn create_claim(&self, vehicle_id: &str) -> AppResult<Claim> {
        self.get_vehicle(vehicle_id)?;
        let claim = Claim::new(vehicle_id.to_string());
        self.store.save_claim(claim.clone());
        Ok(claim)
    }

    pub fn get_claim(&self, id: &str) -> AppResult<Claim> {
        self.store
            .get_claim(id)
            .ok_or_else(|| AppError::ClaimNotFound(id.to_string()))
    }

    pub fn list_claims(&self) -> Vec<Claim> {
        self.store.list_claims()
    }

    pub fn list_claims_by_vehicle(&self, vehicle_id: &str) -> AppResult<Vec<Claim>> {
        self.get_vehicle(vehicle_id)?;
        Ok(self.store.list_claims_by_vehicle(vehicle_id))
    }

    fn is_total_loss_threshold_exceeded(
        total_repair_cost: Decimal,
        actual_value: Decimal,
    ) -> AppResult<bool> {
        if actual_value.is_zero() {
            return Err(AppError::ZeroActualValue);
        }

        let threshold = actual_value * Decimal::from(TOTAL_LOSS_THRESHOLD_PERCENT) / Decimal::from(100);
        Ok(total_repair_cost > threshold)
    }

    pub fn add_repair_item(
        &self,
        claim_id: &str,
        name: String,
        item_type: RepairItemType,
        cost: Decimal,
    ) -> AppResult<Claim> {
        let claim = self.get_claim(claim_id)?;
        let vehicle = self.get_vehicle(&claim.vehicle_id)?;

        if claim.status == ClaimStatus::Settled {
            return Err(AppError::ClaimAlreadySettled);
        }

        if claim.status == ClaimStatus::TotalLoss {
            return Err(AppError::ClaimTotalLoss);
        }

        let item = RepairItem::new(name, item_type, cost);
        let mut new_repair_items = claim.repair_items.clone();
        new_repair_items.push(item);

        let new_total = new_repair_items
            .iter()
            .fold(Decimal::ZERO, |acc, i| acc + i.cost);

        let mut updated_claim = claim.clone();

        if Self::is_total_loss_threshold_exceeded(new_total, vehicle.actual_value)? {
            updated_claim.status = ClaimStatus::TotalLoss;
            updated_claim.repair_items = Vec::new();
        } else {
            updated_claim.repair_items = new_repair_items;
        }

        self.store.save_claim(updated_claim.clone());
        Ok(updated_claim)
    }

    pub fn settle_claim(
        &self,
        claim_id: &str,
        salvage_value: Option<Decimal>,
    ) -> AppResult<Claim> {
        let claim = self.get_claim(claim_id)?;
        let vehicle = self.get_vehicle(&claim.vehicle_id)?;

        if claim.status == ClaimStatus::Settled {
            return Err(AppError::ClaimAlreadySettled);
        }

        let mut updated_claim = claim.clone();

        let payout = match claim.status {
            ClaimStatus::TotalLoss => {
                let salvage = salvage_value.unwrap_or(Decimal::ZERO);
                if salvage > vehicle.actual_value {
                    return Err(AppError::SalvageValueExceedsActualValue);
                }
                let payout = vehicle.actual_value - salvage;
                if payout.is_sign_negative() {
                    return Err(AppError::NegativePayout);
                }
                updated_claim.salvage_value = Some(salvage);
                payout
            }
            ClaimStatus::Pending => claim.total_repair_cost(),
            ClaimStatus::Settled => unreachable!(),
        };

        updated_claim.status = ClaimStatus::Settled;
        updated_claim.payout = Some(payout);

        self.store.save_claim(updated_claim.clone());
        Ok(updated_claim)
    }
}
