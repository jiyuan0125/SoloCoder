use thiserror::Error;

#[derive(Error, Debug)]
pub enum ClaimError {
    #[error("责任比例之和必须为100%，当前为: {0}%")]
    InvalidLiabilityRatioSum(String),
    
    #[error("责任比例必须大于0且不超过100%")]
    InvalidLiabilityRatio,
    
    #[error("案件已结案，无法修改")]
    ClaimClosed,
    
    #[error("案件不存在")]
    ClaimNotFound,
    
    #[error("无效的金额")]
    InvalidAmount,
    
    #[error("保单不存在")]
    PolicyNotFound,
    
    #[error("一方不能重复参与同一案件")]
    DuplicateParty,
}
