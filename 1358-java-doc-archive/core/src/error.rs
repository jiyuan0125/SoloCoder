use thiserror::Error;

#[derive(Error, Debug)]
pub enum ArchiveError {
    #[error("档案不存在: {0}")]
    ArchiveNotFound(String),
    #[error("用户不存在: {0}")]
    UserNotFound(String),
    #[error("借阅记录不存在: {0}")]
    BorrowNotFound(String),
    #[error("预约不存在: {0}")]
    ReservationNotFound(String),
    #[error("档案已被借阅")]
    ArchiveAlreadyBorrowed,
    #[error("档案已被预约")]
    ArchiveAlreadyReserved,
    #[error("用户借阅权限不足")]
    InsufficientPermission,
    #[error("档案当前不允许借阅")]
    ArchiveNotAvailable,
    #[error("用户存在超期未还档案")]
    UserHasOverdueItems,
    #[error("用户借阅权限已被暂停")]
    UserSuspended,
    #[error("续借次数已达上限")]
    RenewalLimitExceeded,
    #[error("机密档案不允许续借")]
    ConfidentialRenewalNotAllowed,
    #[error("预约已过期")]
    ReservationExpired,
    #[error("用户已预约该档案")]
    UserAlreadyReserved,
    #[error("用户已借阅该档案")]
    UserAlreadyBorrowed,
    #[error("审批流程不存在")]
    ApprovalNotFound,
    #[error("审批流程已完成")]
    ApprovalAlreadyCompleted,
    #[error("审批人不匹配")]
    ApproverMismatch,
    #[error("审批顺序错误")]
    ApprovalOrderError,
    #[error("档案损坏赔偿未记录")]
    DamageCompensationRequired,
    #[error("预约者未在3个工作日内领取")]
    ReservationNotClaimed,
    #[error("内部错误: {0}")]
    Internal(String),
}
