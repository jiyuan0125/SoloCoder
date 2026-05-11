use thiserror::Error;

#[derive(Error, Debug)]
pub enum ProcessRouteError {
    #[error("工艺路线不存在: {0}")]
    RouteNotFound(String),

    #[error("版本不存在: {0}")]
    VersionNotFound(String),

    #[error("工序不存在: {0}")]
    ProcessNotFound(String),

    #[error("生产任务不存在: {0}")]
    TaskNotFound(String),

    #[error("存在循环依赖")]
    CircularDependency,

    #[error("前置工序未完成: {0:?}")]
    PredecessorsNotCompleted(Vec<String>),

    #[error("工序不是等待状态，无法开始")]
    ProcessNotWaiting,

    #[error("工序不是进行中状态，无法操作")]
    ProcessNotInProgress,

    #[error("工序已完成，无法修改状态")]
    ProcessAlreadyCompleted,

    #[error("版本 {0} 已存在")]
    VersionAlreadyExists(String),

    #[error("生产任务已绑定版本 {0}，不能切换")]
    TaskVersionLocked(String),

    #[error("关键工序 {0} 暂停，可能影响整体工期")]
    CriticalProcessPaused(String),

    #[error("无效的操作: {0}")]
    InvalidOperation(String),

    #[error("内部错误: {0}")]
    Internal(String),
}

pub type Result<T> = std::result::Result<T, ProcessRouteError>;
