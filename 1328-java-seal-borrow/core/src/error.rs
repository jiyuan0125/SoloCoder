use thiserror::Error;

#[derive(Error, Debug)]
pub enum SealBorrowError {
    #[error("印章不存在")]
    SealNotFound,
    
    #[error("员工不存在")]
    EmployeeNotFound,
    
    #[error("借用申请不存在")]
    BorrowRequestNotFound,
    
    #[error("印章正在维护中，无法借用")]
    SealInMaintenance,
    
    #[error("印章已借出")]
    SealAlreadyBorrowed,
    
    #[error("您已借用该印章，不能同时借用同一个印章")]
    AlreadyBorrowingThisSeal,
    
    #[error("审批人和借用人不能是同一个人")]
    ApproverCannotBeBorrower,
    
    #[error("只有保管人可以审批")]
    OnlyCustodianCanApprove,
    
    #[error("申请状态不支持此操作")]
    InvalidStatusForOperation,
    
    #[error("新的归还日期必须在原日期之后")]
    NewReturnDateMustBeAfterOriginal,
    
    #[error("只有借用人可以申请续借")]
    OnlyBorrowerCanRenew,
    
    #[error("内部错误: {0}")]
    InternalError(String),
}

pub type Result<T> = std::result::Result<T, SealBorrowError>;
