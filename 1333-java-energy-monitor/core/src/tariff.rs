use crate::models::{TariffPeriod, TimeRange};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TariffConfig {
    pub peak_price: f64,
    pub flat_price: f64,
    pub valley_price: f64,
    pub peak_periods: Vec<TimeRange>,
    pub flat_periods: Vec<TimeRange>,
    pub valley_periods: Vec<TimeRange>,
}

impl Default for TariffConfig {
    fn default() -> Self {
        Self::standard()
    }
}

impl TariffConfig {
    pub fn standard() -> Self {
        Self {
            peak_price: 1.2,
            flat_price: 0.8,
            valley_price: 0.4,
            peak_periods: vec![TimeRange::new(10, 12), TimeRange::new(14, 17)],
            flat_periods: vec![
                TimeRange::new(7, 10),
                TimeRange::new(12, 14),
                TimeRange::new(17, 21),
            ],
            valley_periods: vec![TimeRange::new(21, 7)],
        }
    }

    pub fn get_price(&self, period: TariffPeriod) -> f64 {
        match period {
            TariffPeriod::Peak => self.peak_price,
            TariffPeriod::Flat => self.flat_price,
            TariffPeriod::Valley => self.valley_price,
        }
    }

    pub fn get_period_for_hour(&self, hour: u32) -> TariffPeriod {
        for range in &self.peak_periods {
            if range.contains_hour(hour) {
                return TariffPeriod::Peak;
            }
        }
        for range in &self.flat_periods {
            if range.contains_hour(hour) {
                return TariffPeriod::Flat;
            }
        }
        TariffPeriod::Valley
    }
}
