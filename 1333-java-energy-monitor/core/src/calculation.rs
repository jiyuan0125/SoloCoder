use crate::models::{DailyCostBreakdown, ElectricityUsage};
use crate::tariff::TariffConfig;

pub fn calculate_daily_cost(usage: &ElectricityUsage, config: &TariffConfig) -> DailyCostBreakdown {
    let peak_cost = usage.peak * config.peak_price;
    let flat_cost = usage.flat * config.flat_price;
    let valley_cost = usage.valley * config.valley_price;

    DailyCostBreakdown {
        peak_cost,
        flat_cost,
        valley_cost,
        total_cost: peak_cost + flat_cost + valley_cost,
    }
}
