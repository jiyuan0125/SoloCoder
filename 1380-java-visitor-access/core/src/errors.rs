use std::fmt;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum VisitorError {
    VisitorNotFound,
    PasscodeAlreadyUsed,
    PasscodeExpired,
    PasscodeInvalid,
    InvalidStatusTransition,
    VisitorRejected,
    PasscodeGenerationFailed,
    InvalidInput(String),
}

impl fmt::Display for VisitorError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            VisitorError::VisitorNotFound => write!(f, "访客记录未找到"),
            VisitorError::PasscodeAlreadyUsed => write!(f, "通行码已被使用"),
            VisitorError::PasscodeExpired => write!(f, "通行码已过期"),
            VisitorError::PasscodeInvalid => write!(f, "通行码无效"),
            VisitorError::InvalidStatusTransition => write!(f, "无效的状态转换"),
            VisitorError::VisitorRejected => write!(f, "访问已被拒绝，请修改信息后重新提交"),
            VisitorError::PasscodeGenerationFailed => write!(f, "生成通行码失败"),
            VisitorError::InvalidInput(msg) => write!(f, "输入无效: {}", msg),
        }
    }
}

impl std::error::Error for VisitorError {}
