use thiserror::Error;

#[derive(Debug, Error)]
pub enum DrugTraceError {
    #[error("药品不存在: {0}")]
    DrugNotFound(String),
    #[error("批号不存在: {0}")]
    BatchNotFound(String),
    #[error("库存不足")]
    InsufficientStock,
    #[error("数量必须大于0")]
    InvalidQuantity,
    #[error("已过期，不允许出库")]
    Expired,
    #[error("批次已冻结")]
    BatchFrozen,
    #[error("批次已召回")]
    BatchRecalled,
    #[error("药品名称不能为空")]
    EmptyDrugName,
    #[error("批号不能为空")]
    EmptyBatchNo,
    #[error("供应商不能为空")]
    EmptySupplier,
    #[error("日期无效")]
    InvalidDate,
}

pub type Result<T> = std::result::Result<T, DrugTraceError>;
