use thiserror::Error;

#[derive(Error, Debug)]
pub enum SystemError {
    #[error("商品不存在: {0}")]
    ProductNotFound(String),

    #[error("会员不存在: {0}")]
    MemberNotFound(String),

    #[error("订单不存在: {0}")]
    OrderNotFound(String),

    #[error("审批不存在: {0}")]
    ApprovalNotFound(String),

    #[error("库存不足: {0}")]
    InsufficientStock(String),

    #[error("价格计算失败: {0}")]
    PriceCalculationFailed(String),

    #[error("操作被拒绝: {0}")]
    OperationDenied(String),

    #[error("并发冲突: {0}")]
    ConcurrencyConflict(String),
}

pub type Result<T> = std::result::Result<T, SystemError>;
