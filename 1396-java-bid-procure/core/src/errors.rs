use thiserror::Error;

#[derive(Debug, Error)]
pub enum ProcurementError {
    #[error("采购需求不存在: {0}")]
    ProcurementNotFound(String),

    #[error("供应商不存在: {0}")]
    SupplierNotFound(String),

    #[error("无效的状态转换: 当前状态 {current}, 期望操作 {operation}")]
    InvalidStatusTransition {
        current: String,
        operation: String,
    },

    #[error("资质等级不足: 供应商等级 {supplier}, 要求 {required}")]
    InsufficientQualification {
        supplier: String,
        required: String,
    },

    #[error("供应商未被邀请")]
    NotInvited,

    #[error("供应商已弃权，无法继续参与")]
    SupplierWithdrawn,

    #[error("报价必须大于0")]
    InvalidPrice,

    #[error("交货期不能早于发布日期")]
    InvalidDeliveryDate,

    #[error("已在当前轮次提交报价")]
    AlreadySubmitted,

    #[error("不是报价阶段")]
    NotInBiddingPhase,

    #[error("报价不能高于上一轮")]
    PriceNotDecreased,

    #[error("采购需求在报价中，不能修改")]
    CannotModifyDuringBidding,

    #[error("开标供应商不足3家，废标")]
    InsufficientSuppliers,

    #[error("轮次已结束")]
    RoundEnded,

    #[error("未找到有效报价")]
    NoValidBids,
}
