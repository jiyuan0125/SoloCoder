use crate::models::{CargoType, DestinationType, FreightBreakdown, Package};

pub const VOLUMETRIC_DIVISOR: f64 = 6000.0;
pub const REMOTE_SURCHARGE_RATE: f64 = 0.30;
pub const INSURANCE_RATE: f64 = 0.005;
pub const MIN_INSURANCE_FEE: f64 = 5.0;

pub fn calculate_volumetric_weight(length_cm: f64, width_cm: f64, height_cm: f64) -> f64 {
    (length_cm * width_cm * height_cm) / VOLUMETRIC_DIVISOR
}

pub fn calculate_chargeable_weight(volumetric_weight: f64, actual_weight_kg: f64) -> f64 {
    volumetric_weight.max(actual_weight_kg)
}

pub fn calculate_base_freight(chargeable_weight: f64, destination: DestinationType) -> f64 {
    chargeable_weight * destination.price_per_kg()
}

pub fn calculate_remote_surcharge(base_freight: f64, destination: DestinationType) -> f64 {
    if matches!(destination, DestinationType::Remote) {
        base_freight * REMOTE_SURCHARGE_RATE
    } else {
        0.0
    }
}

pub fn calculate_insurance_fee(
    cargo_type: CargoType,
    declared_value: Option<f64>,
) -> f64 {
    if matches!(cargo_type, CargoType::Fragile) {
        let value = declared_value.unwrap_or(0.0);
        let fee = value * INSURANCE_RATE;
        fee.max(MIN_INSURANCE_FEE)
    } else {
        0.0
    }
}

pub fn calculate_freight_breakdown(package: &Package) -> FreightBreakdown {
    let volumetric_weight = calculate_volumetric_weight(
        package.length_cm,
        package.width_cm,
        package.height_cm,
    );
    let chargeable_weight = calculate_chargeable_weight(volumetric_weight, package.actual_weight_kg);
    let base_freight = calculate_base_freight(chargeable_weight, package.destination);
    let remote_surcharge = calculate_remote_surcharge(base_freight, package.destination);
    let insurance_fee = calculate_insurance_fee(package.cargo_type, package.declared_value);
    let total_freight = base_freight + remote_surcharge + insurance_fee;

    FreightBreakdown {
        volumetric_weight,
        chargeable_weight,
        base_freight,
        remote_surcharge,
        insurance_fee,
        total_freight,
    }
}
