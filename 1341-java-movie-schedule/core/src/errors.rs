use thiserror::Error;

#[derive(Debug, Error)]
pub enum ScheduleError {
    #[error("时间冲突：与排片 {0} 重叠")]
    Conflict(String),
    
    #[error("黄金时段排片超过限制（每天最多2场）")]
    PrimeTimeLimitExceeded,
    
    #[error("影厅不存在")]
    HallNotFound,
    
    #[error("影片不存在")]
    MovieNotFound,
    
    #[error("排片不存在")]
    ScheduleNotFound,
}
