use thiserror::Error;

#[derive(Debug, Error)]
pub enum SchedulerError {
    #[error("患者不存在: {0}")]
    PatientNotFound(String),

    #[error("机器不存在: {0}")]
    MachineNotFound(String),

    #[error("排班记录不存在: {0}")]
    ScheduleNotFound(String),

    #[error("维护记录不存在: {0}")]
    MaintenanceNotFound(String),

    #[error("机器与患者传染病类型不匹配")]
    DiseaseMismatch,

    #[error("乙肝患者必须使用乙肝专用机")]
    HepatitisBMachineRequired,

    #[error("非乙肝患者不能使用乙肝专用机")]
    NonHepatitisBCannotUseHepatitisBMachine,

    #[error("同一时段机器已被占用")]
    MachineAlreadyOccupied,

    #[error("患者当天已有排班")]
    PatientAlreadyScheduledOnDay,

    #[error("机器在该时段处于维护中")]
    MachineUnderMaintenance,

    #[error("批量排班存在冲突")]
    BulkScheduleConflict,

    #[error("没有可用的替代机器")]
    NoAlternativeMachineAvailable,

    #[error("未找到空闲机器")]
    NoIdleMachineFound,

    #[error("调班目标时段不可用")]
    RescheduleTargetNotAvailable,
}
