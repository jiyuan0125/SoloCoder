use thiserror::Error;

#[derive(Error, Debug)]
pub enum AppError {
    #[error("部门不存在: {0}")]
    DepartmentNotFound(String),
    
    #[error("员工不存在: {0}")]
    EmployeeNotFound(String),
    
    #[error("员工已存在: {0}")]
    EmployeeAlreadyExists(String),
    
    #[error("岗位不存在: {0}")]
    PositionNotFound(String),
    
    #[error("评估记录不存在: {0}")]
    AssessmentNotFound(String),
    
    #[error("该员工在{0}年已有评估记录，评分不可修改，如需调整请创建新的评估记录")]
    AssessmentAlreadyExistsForYear(i32),
    
    #[error("评分无效: 绩效和潜力都必须在1-5之间")]
    InvalidRating,
    
    #[error("部门数据隔离违规")]
    DataIsolationViolation,
    
    #[error("内部错误: {0}")]
    Internal(String),
}
