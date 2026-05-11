use std::fmt;
use serde::{Serialize, Deserialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ExpressError {
    PackageNotFound,
    PackageAlreadyPickedUp,
    PackageLocked,
    PickupCodeIncorrect,
    VerificationFailed,
    AlreadyProxy,
    InvalidPhone,
    InvalidTrackingNumber,
    InternalError(String),
}

impl fmt::Display for ExpressError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ExpressError::PackageNotFound => write!(f, "包裹不存在"),
            ExpressError::PackageAlreadyPickedUp => write!(f, "包裹已被取走"),
            ExpressError::PackageLocked => write!(f, "包裹已锁定，30分钟后重试或凭手机号和身份证后4位核验"),
            ExpressError::PickupCodeIncorrect => write!(f, "取件码错误"),
            ExpressError::VerificationFailed => write!(f, "人工核验失败"),
            ExpressError::AlreadyProxy => write!(f, "该包裹已登记代收"),
            ExpressError::InvalidPhone => write!(f, "手机号格式无效"),
            ExpressError::InvalidTrackingNumber => write!(f, "运单号无效"),
            ExpressError::InternalError(msg) => write!(f, "内部错误: {}", msg),
        }
    }
}

impl std::error::Error for ExpressError {}
