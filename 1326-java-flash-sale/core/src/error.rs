use thiserror::Error;

#[derive(Debug, Error)]
pub enum FlashSaleError {
    #[error("活动不存在")]
    ActivityNotFound,
    
    #[error("活动还未开始")]
    ActivityNotStarted,
    
    #[error("活动已结束")]
    ActivityEnded,
    
    #[error("库存不足")]
    OutOfStock,
    
    #[error("用户已成功购买过该活动商品")]
    UserAlreadyPurchased,
    
    #[error("订单不存在")]
    OrderNotFound,
    
    #[error("订单状态不允许此操作")]
    InvalidOrderStatus,
    
    #[error("系统错误: {0}")]
    InternalError(String),
}

pub type Result<T> = std::result::Result<T, FlashSaleError>;
