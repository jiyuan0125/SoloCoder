use serde::{Serialize, Deserialize};
use std::collections::HashSet;

#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum ServiceType {
    A,
    B,
    C,
}

impl ServiceType {
    pub fn prefix(&self) -> &str {
        match self {
            ServiceType::A => "A",
            ServiceType::B => "B",
            ServiceType::C => "C",
        }
    }
}

impl std::fmt::Display for ServiceType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.prefix())
    }
}

impl std::str::FromStr for ServiceType {
    type Err = String;
    
    fn from_str(s: &str) -> Result<Self, Self::Err> {
        match s.to_uppercase().as_str() {
            "A" => Ok(ServiceType::A),
            "B" => Ok(ServiceType::B),
            "C" => Ok(ServiceType::C),
            _ => Err(format!("Invalid service type: {}", s)),
        }
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct TicketNumber {
    pub service_type: ServiceType,
    pub number: u32,
    pub is_vip: bool,
}

impl TicketNumber {
    pub fn new(service_type: ServiceType, number: u32, is_vip: bool) -> Self {
        Self { service_type, number, is_vip }
    }
    
    pub fn display(&self) -> String {
        let vip_prefix = if self.is_vip { "VIP-" } else { "" };
        format!("{}{}{:03}", vip_prefix, self.service_type.prefix(), self.number)
    }
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum TicketStatus {
    Waiting,
    Called,
    Processing,
    Missed,
    Completed,
    Cancelled,
    Expired,
}

#[derive(Debug, Clone)]
pub struct Ticket {
    pub number: TicketNumber,
    pub status: TicketStatus,
    pub call_count: u32,
    pub missed_count: u32,
    pub created_at: std::time::Instant,
    pub called_at: Option<std::time::Instant>,
}

impl Ticket {
    pub fn new(number: TicketNumber) -> Self {
        Self {
            number,
            status: TicketStatus::Waiting,
            call_count: 0,
            missed_count: 0,
            created_at: std::time::Instant::now(),
            called_at: None,
        }
    }
    
    pub fn can_be_queued(&self) -> bool {
        self.missed_count < 2
    }
    
    pub fn loses_vip_priority(&self) -> bool {
        self.missed_count > 0
    }
    
    pub fn to_dto(&self) -> TicketDto {
        TicketDto {
            number: self.number.clone(),
            status: self.status.clone(),
            call_count: self.call_count,
            missed_count: self.missed_count,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TicketDto {
    pub number: TicketNumber,
    pub status: TicketStatus,
    pub call_count: u32,
    pub missed_count: u32,
}

#[derive(Debug, Clone)]
pub struct Window {
    pub id: u32,
    pub name: String,
    pub service_types: HashSet<ServiceType>,
    pub is_active: bool,
    pub current_ticket: Option<Ticket>,
}

impl Window {
    pub fn new(id: u32, name: String, service_types: HashSet<ServiceType>) -> Self {
        Self {
            id,
            name,
            service_types,
            is_active: true,
            current_ticket: None,
        }
    }
    
    pub fn can_handle(&self, service_type: &ServiceType) -> bool {
        self.is_active && self.service_types.contains(service_type)
    }
    
    pub fn is_free(&self) -> bool {
        self.is_active && self.current_ticket.is_none()
    }
    
    pub fn to_dto(&self) -> WindowDto {
        WindowDto {
            id: self.id,
            name: self.name.clone(),
            service_types: self.service_types.clone(),
            is_active: self.is_active,
            current_ticket: self.current_ticket.as_ref().map(|t| t.to_dto()),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct WindowDto {
    pub id: u32,
    pub name: String,
    pub service_types: HashSet<ServiceType>,
    pub is_active: bool,
    pub current_ticket: Option<TicketDto>,
}

#[derive(Debug, Clone)]
pub struct CallRecord {
    pub ticket_number: TicketNumber,
    pub window_id: u32,
    pub call_count: u32,
    pub called_at: std::time::Instant,
}
