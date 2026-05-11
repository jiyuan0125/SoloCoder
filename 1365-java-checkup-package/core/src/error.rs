use thiserror::Error;

#[derive(Error, Debug)]
pub enum CheckupError {
    #[error("项目不存在: {0}")]
    ItemNotFound(String),
    
    #[error("套餐不存在: {0}")]
    PackageNotFound(String),
    
    #[error("客户不存在: {0}")]
    CustomerNotFound(String),
    
    #[error("互斥规则冲突: {0}")]
    HardMutexConflict(String),
    
    #[error("性别限制冲突: {0}")]
    GenderRestrictionViolation(String),
    
    #[error("年龄限制冲突: {0}")]
    AgeRestrictionViolation(String),
    
    #[error("孕妇禁忌: {0}")]
    PregnantForbidden(String),
}
