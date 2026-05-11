use thiserror::Error;

#[derive(Error, Debug)]
pub enum BookingError {
    #[error("预订不存在")]
    BookingNotFound,
    #[error("菜单套系不存在")]
    MenuSetNotFound,
    #[error("菜品不存在")]
    MenuItemNotFound,
    #[error("预订已取消或已完成")]
    BookingInactive,
    #[error("日期不能早于当前时间")]
    InvalidEventDate,
    #[error("桌数必须大于0")]
    InvalidTableCount,
    #[error("定金金额无效")]
    InvalidDeposit,
    #[error("菜品替换差价超过原菜品价格的30%")]
    PriceDifferenceExceeded,
    #[error("菜品分类不同，不能替换")]
    DifferentCategory,
    #[error("未用菜品处理方式不一致，不能混合使用")]
    MixedUnusedOption,
    #[error("实际桌数不能大于预订桌数")]
    ActualTablesExceedBooked,
    #[error("操作时间过晚，无法执行该操作")]
    OperationTooLate,
    #[error("金额计算错误")]
    CalculationError,
}
