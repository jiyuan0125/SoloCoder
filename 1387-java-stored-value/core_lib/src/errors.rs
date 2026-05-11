use thiserror::Error;

#[derive(Error, Debug)]
pub enum StoredValueError {
    #[error("会员不存在: {0}")]
    MemberNotFound(String),
    #[error("余额不足，需要: {0}, 可用: {1}")]
    InsufficientBalance(u64, u64),
    #[error("消费记录不存在: {0}")]
    TransactionNotFound(String),
    #[error("退款金额超过消费金额: 退款 {0}, 原始消费 {1}")]
    RefundAmountExceeds(u64, u64),
    #[error("充值金额必须大于0")]
    InvalidRechargeAmount,
    #[error("消费金额必须大于0")]
    InvalidConsumeAmount,
    #[error("内部错误: {0}")]
    InternalError(String),
}

pub type Result<T> = std::result::Result<T, StoredValueError>;
