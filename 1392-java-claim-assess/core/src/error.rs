use thiserror::Error;

#[derive(Debug, Error)]
pub enum AppError {
    #[error("车辆不存在: {0}")]
    VehicleNotFound(String),

    #[error("理赔单不存在: {0}")]
    ClaimNotFound(String),

    #[error("车辆存在未结案的理赔单，不能修改实际价值")]
    HasUnsettledClaim,

    #[error("残值不能超过车辆实际价值")]
    SalvageValueExceedsActualValue,

    #[error("赔付款不能为负")]
    NegativePayout,

    #[error("车辆实际价值为0，无法计算80%阈值")]
    ZeroActualValue,

    #[error("理赔单已结案，不能添加维修项目")]
    ClaimAlreadySettled,

    #[error("理赔单已判定推定全损，不能添加维修项目")]
    ClaimTotalLoss,
}

pub type AppResult<T> = Result<T, AppError>;
