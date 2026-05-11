use std::fmt;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum CheckinError {
    ActivityNotFound,
    EmployeeNotFound,
    ActivityNotStarted,
    ActivityEnded,
    LocationOutOfRange,
    DeviceAlreadyUsedByOther,
    AlreadyCheckedIn,
    NotCheckedIn,
    AlreadyCheckedOut,
    EarlyCheckout,
    InternalError(String),
}

impl fmt::Display for CheckinError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            CheckinError::ActivityNotFound => write!(f, "活动不存在"),
            CheckinError::EmployeeNotFound => write!(f, "员工不存在"),
            CheckinError::ActivityNotStarted => write!(f, "活动尚未开始"),
            CheckinError::ActivityEnded => write!(f, "活动已结束"),
            CheckinError::LocationOutOfRange => write!(f, "不在签到范围内"),
            CheckinError::DeviceAlreadyUsedByOther => write!(f, "该设备已为其他员工签到"),
            CheckinError::AlreadyCheckedIn => write!(f, "已签到，无需重复签到"),
            CheckinError::NotCheckedIn => write!(f, "尚未签到"),
            CheckinError::AlreadyCheckedOut => write!(f, "已签退"),
            CheckinError::EarlyCheckout => write!(f, "签退早于活动结束，记为早退"),
            CheckinError::InternalError(msg) => write!(f, "内部错误: {}", msg),
        }
    }
}

impl std::error::Error for CheckinError {}
