use thiserror::Error;

#[derive(Error, Debug)]
pub enum AssetError {
    #[error("资产不存在: {0}")]
    AssetNotFound(String),

    #[error("资产分类不存在: {0}")]
    CategoryNotFound(String),

    #[error("员工不存在: {0}")]
    EmployeeNotFound(String),

    #[error("领用申请不存在: {0}")]
    ApplicationNotFound(String),

    #[error("资产状态错误: {0}")]
    InvalidAssetStatus(String),

    #[error("员工持有同分类资产已达上限: 分类={category}, 上限={limit}")]
    CategoryLimitExceeded { category: String, limit: u32 },

    #[error("资产当前状态不可领用: {0}")]
    AssetNotAvailable(String),

    #[error("资产已被他人成功申请")]
    ConcurrentReservationFailed,

    #[error("审批人不是管理员")]
    NotAnAdmin,

    #[error("当前用户不是资产责任人，无法转移")]
    NotAssetOwner,

    #[error("转移需要双方确认")]
    TransferNeedsBothParties,

    #[error("归还需要管理员确认")]
    ReturnNeedsAdmin,

    #[error("申请已被处理，无法重复操作")]
    ApplicationAlreadyProcessed,

    #[error("资产当前状态无法进行此操作")]
    InvalidOperation,

    #[error("内部错误: {0}")]
    InternalError(String),
}
