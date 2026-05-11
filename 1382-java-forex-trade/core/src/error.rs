use thiserror::Error;

#[derive(Debug, Error)]
pub enum ForexError {
    #[error("账户不存在: {0}")]
    AccountNotFound(String),
    
    #[error("货币对不存在: {0}")]
    CurrencyPairNotFound(String),
    
    #[error("订单不存在: {0}")]
    OrderNotFound(String),
    
    #[error("余额不足")]
    InsufficientBalance,
    
    #[error("无效的金额")]
    InvalidAmount,
    
    #[error("无效的货币")]
    InvalidCurrency,
    
    #[error("订单状态不允许此操作")]
    InvalidOrderStatus,
    
    #[error("需要审批")]
    ApprovalRequired,
    
    #[error("订单已被拒绝")]
    OrderRejected,
    
    #[error("并发冲突，请重试")]
    ConcurrentConflict,
}

pub type Result<T> = std::result::Result<T, ForexError>;
