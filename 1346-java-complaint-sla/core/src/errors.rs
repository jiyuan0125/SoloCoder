use thiserror::Error;

#[derive(Debug, Error)]
pub enum ComplaintError {
    #[error("投诉不存在: {0}")]
    NotFound(String),
    
    #[error("投诉状态无效，无法执行该操作")]
    InvalidState,
    
    #[error("处理人不存在: {0}")]
    HandlerNotFound(String),
    
    #[error("内部错误: {0}")]
    Internal(String),
}

pub type Result<T> = std::result::Result<T, ComplaintError>;
