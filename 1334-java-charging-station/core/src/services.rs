use chrono::{DateTime, Duration, Timelike, Utc};
use rust_decimal::Decimal;
use rust_decimal::prelude::*;
use uuid::Uuid;

use crate::errors::ChargingError;
use crate::models::{
    Charger, ChargerStatus, Order, OrderStatus, RateConfig, RatePeriod,
};
use crate::storage::SharedState;

pub fn get_rate_period(time: &DateTime<Utc>) -> RatePeriod {
    let hour = time.hour();
    
    match hour {
        7..=11 | 17..=21 => RatePeriod::Peak,
        12..=16 | 22..=23 => RatePeriod::Normal,
        _ => RatePeriod::Valley,
    }
}

pub fn get_rate_for_period(period: RatePeriod, config: &RateConfig) -> Decimal {
    match period {
        RatePeriod::Peak => config.peak_rate,
        RatePeriod::Normal => config.normal_rate,
        RatePeriod::Valley => config.valley_rate,
    }
}

pub fn calculate_minutes(
    start: DateTime<Utc>,
    end: DateTime<Utc>,
) -> u64 {
    let duration = end.signed_duration_since(start);
    let mut minutes = duration.num_minutes();
    
    if minutes < 0 {
        minutes = 0;
    }
    
    if minutes == 0 && duration.num_seconds() > 0 {
        minutes = 1;
    }
    
    minutes as u64
}

pub fn calculate_energy(
    power_kw: u32,
    minutes: u64,
) -> Decimal {
    let hours = Decimal::from(minutes) / Decimal::from(60);
    Decimal::from(power_kw) * hours
}

pub struct Segment {
    pub start: DateTime<Utc>,
    pub end: DateTime<Utc>,
    pub period: RatePeriod,
    pub minutes: u64,
    pub energy: Decimal,
    pub rate: Decimal,
    pub cost: Decimal,
}

pub fn calculate_segments(
    start: DateTime<Utc>,
    end: DateTime<Utc>,
    power_kw: u32,
    config: &RateConfig,
) -> Vec<Segment> {
    let mut segments = Vec::new();
    let mut current = start;
    
    while current < end {
        let current_period = get_rate_period(&current);
        let next_hour_start = get_next_hour_boundary(current);
        let segment_end = if next_hour_start < end { next_hour_start } else { end };
        let minutes = calculate_minutes(current, segment_end);
        let energy = calculate_energy(power_kw, minutes);
        let rate = get_rate_for_period(current_period, config);
        let cost = energy * rate;
        
        segments.push(Segment {
            start: current,
            end: segment_end,
            period: current_period,
            minutes,
            energy,
            rate,
            cost,
        });
        
        current = segment_end;
    }
    
    segments
}

fn get_next_hour_boundary(time: DateTime<Utc>) -> DateTime<Utc> {
    let next_hour = time.hour() + 1;
    if next_hour >= 24 {
        let next_day = time.date_naive() + Duration::days(1);
        next_day.and_hms_opt(0, 0, 0).unwrap().and_utc()
    } else {
        time.date_naive().and_hms_opt(next_hour, 0, 0).unwrap().and_utc()
    }
}

pub fn calculate_total_cost(segments: &[Segment]) -> Decimal {
    let mut total = Decimal::ZERO;
    for segment in segments {
        total += segment.cost;
    }
    
    let min_charge = Decimal::from_f64(0.01).unwrap();
    if total > Decimal::ZERO && total < min_charge {
        min_charge
    } else {
        total
    }
}

pub async fn create_order(
    state: &SharedState,
    charger_id: Uuid,
    user_id: Uuid,
) -> Result<Order, ChargingError> {
    let mut state_guard = state.lock().await;
    
    let charger = state_guard.chargers.get(&charger_id)
        .cloned()
        .ok_or_else(|| ChargingError::ChargerNotFound(charger_id.to_string()))?;
    
    if charger.status == ChargerStatus::Charging {
        return Err(ChargingError::ChargerBusy(charger_id.to_string()));
    }
    
    let order = Order {
        id: Uuid::new_v4(),
        charger_id,
        user_id,
        start_time: Utc::now(),
        end_time: None,
        duration_minutes: None,
        energy_kwh: None,
        total_cost: None,
        status: OrderStatus::Active,
    };
    
    state_guard.orders.insert(order.id, order.clone());
    
    let mut updated_charger = charger;
    updated_charger.status = ChargerStatus::Charging;
    state_guard.chargers.insert(charger_id, updated_charger);
    
    drop(state_guard);
    
    Ok(order)
}

pub async fn end_order(
    state: &SharedState,
    order_id: Uuid,
) -> Result<Order, ChargingError> {
    let mut state_guard = state.lock().await;
    
    let order = state_guard.orders.get(&order_id)
        .cloned()
        .ok_or_else(|| ChargingError::OrderNotFound(order_id.to_string()))?;
    
    if order.status == OrderStatus::Completed {
        return Err(ChargingError::OrderAlreadyCompleted(order_id.to_string()));
    }
    
    let charger = state_guard.chargers.get(&order.charger_id)
        .cloned()
        .ok_or_else(|| ChargingError::ChargerNotFound(order.charger_id.to_string()))?;
    
    let end_time = Utc::now();
    let duration_minutes = calculate_minutes(order.start_time, end_time);
    let config = state_guard.rate_config;
    
    let segments = calculate_segments(order.start_time, end_time, charger.power_kw, &config);
    let total_energy: Decimal = segments.iter().map(|s| s.energy).sum();
    let total_cost = calculate_total_cost(&segments);
    
    let mut updated_order = order;
    updated_order.end_time = Some(end_time);
    updated_order.duration_minutes = Some(duration_minutes);
    updated_order.energy_kwh = Some(total_energy);
    updated_order.total_cost = Some(total_cost);
    updated_order.status = OrderStatus::Completed;
    
    let charger_id = updated_order.charger_id;
    state_guard.orders.insert(order_id, updated_order.clone());
    
    let mut updated_charger = charger;
    updated_charger.status = ChargerStatus::Idle;
    state_guard.chargers.insert(charger_id, updated_charger);
    
    drop(state_guard);
    
    Ok(updated_order)
}

pub async fn create_charger(
    state: &SharedState,
    name: String,
    power_kw: u32,
) -> Charger {
    let charger = Charger {
        id: Uuid::new_v4(),
        name,
        power_kw,
        status: ChargerStatus::Idle,
    };
    
    let mut state_guard = state.lock().await;
    state_guard.chargers.insert(charger.id, charger.clone());
    drop(state_guard);
    
    charger
}
