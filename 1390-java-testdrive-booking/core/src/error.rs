use thiserror::Error;

#[derive(Debug, Error)]
pub enum BookingError {
    #[error("车型不存在")]
    CarModelNotFound,
    #[error("没有可用的试驾车")]
    NoAvailableCar,
    #[error("该时段已无可用车辆，请换时段")]
    TimeSlotUnavailable,
    #[error("您今天已经预约过了")]
    AlreadyBookedToday,
    #[error("预约不存在")]
    BookingNotFound,
    #[error("销售顾问不存在")]
    AdvisorNotFound,
    #[error("无效的预约状态转换")]
    InvalidStateTransition,
    #[error("预约已过期")]
    BookingExpired,
    #[error("内部错误: {0}")]
    Internal(String),
}

pub type Result<T> = std::result::Result<T, BookingError>;
