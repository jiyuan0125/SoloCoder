use serde::{Serialize, Deserialize};
use std::fmt;
use crate::models::ServiceType;

#[derive(Debug)]
pub enum QueueError {
    InvalidServiceType(String),
    WindowNotFound(u32),
    NoAvailableWindow(ServiceType),
    NoWaitingTickets,
    TicketNotFound(String),
    WindowAlreadyClosed(u32),
    WindowAlreadyOpen(u32),
    WindowBusy(u32),
    TooManyMisses(String),
    InvalidTicketStatus(String, String),
}

impl fmt::Display for QueueError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            QueueError::InvalidServiceType(s) => write!(f, "Invalid service type: {}", s),
            QueueError::WindowNotFound(id) => write!(f, "Window not found: {}", id),
            QueueError::NoAvailableWindow(st) => write!(f, "No available window for service type: {}", st),
            QueueError::NoWaitingTickets => write!(f, "No waiting tickets"),
            QueueError::TicketNotFound(s) => write!(f, "Ticket not found: {}", s),
            QueueError::WindowAlreadyClosed(id) => write!(f, "Window already closed: {}", id),
            QueueError::WindowAlreadyOpen(id) => write!(f, "Window already open: {}", id),
            QueueError::WindowBusy(id) => write!(f, "Window is busy: {}", id),
            QueueError::TooManyMisses(s) => write!(f, "Ticket has too many misses: {}", s),
            QueueError::InvalidTicketStatus(s, msg) => write!(f, "Invalid ticket status for {}: {}", s, msg),
        }
    }
}

impl std::error::Error for QueueError {}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ErrorResponse {
    pub error: String,
}
