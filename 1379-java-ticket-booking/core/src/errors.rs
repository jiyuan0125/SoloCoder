use thiserror::Error;

#[derive(Error, Debug)]
pub enum TicketError {
    #[error("景区不存在")]
    ScenicNotFound,
    
    #[error("门票类型无效")]
    InvalidTicketType,
    
    #[error("身份证格式无效")]
    InvalidIdCard,
    
    #[error("该身份证当天已购买过门票")]
    IdCardAlreadyPurchased,
    
    #[error("超出景区当日最大承载量")]
    CapacityExceeded,
    
    #[error("免票儿童数量已达上限（每张成人票限带1个免票儿童）")]
    FreeChildrenExceeded,
    
    #[error("无效的使用日期")]
    InvalidDate,
    
    #[error("退票时间已过（需在使用日期前一天24点前退票）")]
    RefundTimeExceeded,
    
    #[error("订单不存在")]
    OrderNotFound,
    
    #[error("订单已退票")]
    OrderAlreadyRefunded,
    
    #[error("并发冲突，请重试")]
    ConcurrencyConflict,
    
    #[error("内部错误: {0}")]
    Internal(String),
    
    #[error("参数错误: {0}")]
    InvalidParam(String),
}

pub type Result<T> = std::result::Result<T, TicketError>;
