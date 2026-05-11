use thiserror::Error;

#[derive(Error, Debug, Clone)]
pub enum ReviewError {
    #[error("评价不存在")]
    ReviewNotFound,
    
    #[error("追评窗口已关闭（超过7天）")]
    FollowUpWindowClosed,
    
    #[error("已经发表过追评")]
    FollowUpAlreadyExists,
    
    #[error("商家已经回复过此评价")]
    ReplyAlreadyExists,
    
    #[error("星级必须在1-5之间")]
    InvalidRating,
    
    #[error("评价内容不能为空")]
    EmptyContent,
    
    #[error("商家没有权限删除评价")]
    PermissionDenied,
    
    #[error("商品ID不能为空")]
    EmptyProductId,
    
    #[error("顾客ID不能为空")]
    EmptyCustomerId,
}
