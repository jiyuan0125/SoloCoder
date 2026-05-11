use crate::models::{CargoType, CreatePackageRequest, DestinationType, Order};
use chrono::Utc;
use thiserror::Error;

#[derive(Debug, Error, Clone, Serialize)]
pub enum ValidationError {
    #[error("包裹尺寸无效：长、宽、高必须大于0")]
    InvalidDimensions,
    #[error("重量无效：实际重量必须大于0")]
    InvalidWeight,
    #[error("声明价值无效：必须大于等于0")]
    InvalidDeclaredValue,
    #[error("液体货物不能发往偏远地区")]
    LiquidToRemoteNotAllowed,
    #[error("易碎品必须提供声明价值")]
    FragileRequiresDeclaredValue,
}

use serde::Serialize;

pub fn validate_package(pkg: &CreatePackageRequest) -> Result<(), ValidationError> {
    if pkg.length_cm <= 0.0 || pkg.width_cm <= 0.0 || pkg.height_cm <= 0.0 {
        return Err(ValidationError::InvalidDimensions);
    }

    if pkg.actual_weight_kg <= 0.0 {
        return Err(ValidationError::InvalidWeight);
    }

    if let Some(declared_value) = pkg.declared_value {
        if declared_value < 0.0 {
            return Err(ValidationError::InvalidDeclaredValue);
        }
    }

    if matches!(pkg.cargo_type, CargoType::Liquid)
        && matches!(pkg.destination, DestinationType::Remote)
    {
        return Err(ValidationError::LiquidToRemoteNotAllowed);
    }

    if matches!(pkg.cargo_type, CargoType::Fragile) && pkg.declared_value.is_none() {
        return Err(ValidationError::FragileRequiresDeclaredValue);
    }

    Ok(())
}

pub fn can_cancel_order(order: &Order) -> bool {
    if order.status != crate::models::OrderStatus::Created {
        return false;
    }

    let now = Utc::now();
    let elapsed = now.signed_duration_since(order.created_at);
    elapsed.num_hours() < 24
}

#[derive(Debug, Error, Clone, Serialize)]
pub enum CancelError {
    #[error("订单不是可取消状态（当前状态：{0}）")]
    NotCancellable(String),
    #[error("订单已超过24小时取消期限")]
    CancelPeriodExpired,
    #[error("订单不存在")]
    OrderNotFound,
}
