use thiserror::Error;

#[derive(Debug, Error)]
pub enum HousekeepingError {
    #[error("阿姨不存在: {0}")]
    AuntNotFound(String),
    
    #[error("客户不存在: {0}")]
    CustomerNotFound(String),
    
    #[error("订单不存在: {0}")]
    OrderNotFound(String),
    
    #[error("退款申请不存在: {0}")]
    RefundRequestNotFound(String),
    
    #[error("阿姨不具备该技能")]
    AuntNoSkill,
    
    #[error("阿姨该时段不可用")]
    AuntNotAvailable,
    
    #[error("阿姨该时段已有冲突订单")]
    AuntTimeConflict,
    
    #[error("没有可用的阿姨")]
    NoAvailableAunt,
    
    #[error("订单状态不正确: 当前状态 {current}, 期望状态 {expected}")]
    InvalidOrderStatus { current: String, expected: String },
    
    #[error("退款比例不能超过50%")]
    RefundPercentageExceeded,
    
    #[error("退款比例必须在0-100之间")]
    InvalidRefundPercentage,
    
    #[error("只有订单所有者可以申请退款")]
    RefundNotOwner,
    
    #[error("时段无效: 结束时间必须晚于开始时间")]
    InvalidTimeSlot,
    
    #[error("并发冲突: 订单已被其他客户预约")]
    ConcurrentBookingConflict,
    
    #[error("时薪必须大于0")]
    InvalidHourlyRate,
    
    #[error("电话号码不能为空")]
    PhoneNumberEmpty,
    
    #[error("姓名不能为空")]
    NameEmpty,
    
    #[error("地址不能为空")]
    AddressEmpty,
    
    #[error("阿姨至少需要具备一个技能")]
    AtLeastOneSkillRequired,
    
    #[error("阿姨至少需要设置一个可用时段")]
    AtLeastOneAvailableTimeRequired,
    
    #[error("内部错误: {0}")]
    Internal(String),
}

pub type Result<T> = std::result::Result<T, HousekeepingError>;
