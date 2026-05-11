use thiserror::Error;

#[derive(Error, Debug)]
pub enum LeaveError {
    #[error("员工不存在: {0}")]
    EmployeeNotFound(String),
    #[error("申请不存在: {0}")]
    LeaveNotFound(String),
    #[error("假期额度不足，需要 {need} 天，可用 {available} 天")]
    InsufficientLeaveBalance { need: u32, available: u32 },
    #[error("已开始的假期不能撤销")]
    LeaveAlreadyStarted,
    #[error("病假超过3天需要提供证明")]
    SickLeaveNeedsProof,
    #[error("日期无效: {0}")]
    InvalidDate(String),
    #[error("年假计算错误: {0}")]
    AnnualLeaveCalculationError(String),
}

pub type Result<T> = std::result::Result<T, LeaveError>;
