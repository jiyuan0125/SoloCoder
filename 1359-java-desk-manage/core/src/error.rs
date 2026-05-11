use thiserror::Error;

#[derive(Debug, Error)]
pub enum DeskError {
    #[error("工位不存在: {0}")]
    DeskNotFound(String),
    
    #[error("员工不存在: {0}")]
    EmployeeNotFound(String),
    
    #[error("部门不存在: {0}")]
    DepartmentNotFound(String),
    
    #[error("预约不存在: {0}")]
    ReservationNotFound(String),
    
    #[error("没有可用的固定工位")]
    NoAvailableFixedDesk,
    
    #[error("工位已被占用")]
    DeskAlreadyOccupied,
    
    #[error("员工当天已有预约")]
    EmployeeAlreadyHasReservation,
    
    #[error("预约时间超出范围（最多提前5个工作日）")]
    ReservationTimeOutOfRange,
    
    #[error("预约已过期")]
    ReservationExpired,
    
    #[error("预约已签到")]
    ReservationAlreadyCheckedIn,
    
    #[error("批量操作冲突: {0}")]
    BatchConflict(String),
    
    #[error("并发分配冲突")]
    ConcurrentAllocationConflict,
    
    #[error("无效操作: {0}")]
    InvalidOperation(String),
}

pub type Result<T> = std::result::Result<T, DeskError>;
