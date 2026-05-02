use std::fmt;

#[derive(Debug, PartialEq, Eq)]
pub enum QueueError {
    Full,
    Empty,
    Closed,
}

impl fmt::Display for QueueError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            QueueError::Full => write!(f, "Queue is full"),
            QueueError::Empty => write!(f, "Queue is empty"),
            QueueError::Closed => write!(f, "Queue is closed"),
        }
    }
}

impl std::error::Error for QueueError {}

pub type QueueResult<T> = Result<T, QueueError>;
