use thiserror::Error;

#[derive(Debug, Error)]
pub enum InvoiceError {
    #[error("发票号码已存在 (代码: {0}, 号码: {1})")]
    InvoiceNumberExists(String, String),

    #[error("开票日期不能是未来日期")]
    FutureDate,

    #[error("发票不存在: {0}")]
    InvoiceNotFound(String),

    #[error("发票已作废，不能进行此操作")]
    AlreadyVoided,

    #[error("发票已红冲，不能进行此操作")]
    AlreadyRed冲ed,

    #[error("红冲发票不能再次红冲或作废")]
    Red冲InvoiceCannotBeOperated,

    #[error("金额必须大于0")]
    AmountMustBePositive,

    #[error("税率必须在0到1之间")]
    InvalidTaxRate,

    #[error("内部错误: {0}")]
    Internal(String),
}
