use thiserror::Error;

#[derive(Error, Debug)]
pub enum StockError {
    #[error("物料不存在: {0}")]
    MaterialNotFound(String),

    #[error("盘点批次不存在: {0}")]
    BatchNotFound(String),

    #[error("差异不存在: {0}")]
    DifferenceNotFound(String),

    #[error("批次状态不允许该操作: {0}")]
    InvalidBatchState(String),

    #[error("差异状态不允许该操作: {0}")]
    InvalidDifferenceState(String),

    #[error("原因说明不能为空")]
    ReasonRequired,

    #[error("并发冲突，请重试")]
    ConcurrencyConflict,

    #[error("内部错误: {0}")]
    Internal(String),
}

pub type StockResult<T> = Result<T, StockError>;
