use std::collections::{HashMap, VecDeque, HashSet};

use crate::models::{
    ServiceType, TicketNumber, Ticket, TicketStatus, Window, CallRecord
};
use crate::errors::QueueError;

#[derive(Debug, Clone)]
pub struct QueueManager {
    counters: HashMap<ServiceType, u32>,
    vip_counters: HashMap<ServiceType, u32>,
    windows: HashMap<u32, Window>,
    service_queues: HashMap<ServiceType, VecDeque<Ticket>>,
    all_tickets: HashMap<String, Ticket>,
    call_history: Vec<CallRecord>,
    next_window_id: u32,
}

impl QueueManager {
    pub fn new() -> Self {
        Self {
            counters: HashMap::new(),
            vip_counters: HashMap::new(),
            windows: HashMap::new(),
            service_queues: HashMap::new(),
            all_tickets: HashMap::new(),
            call_history: Vec::new(),
            next_window_id: 1,
        }
    }

    pub fn add_window(&mut self, name: String, service_types: HashSet<ServiceType>) -> Window {
        let id = self.next_window_id;
        self.next_window_id += 1;
        let window = Window::new(id, name, service_types);
        self.windows.insert(id, window.clone());
        window
    }

    pub fn get_window(&self, id: u32) -> Option<&Window> {
        self.windows.get(&id)
    }

    pub fn get_all_windows(&self) -> Vec<&Window> {
        self.windows.values().collect()
    }

    pub fn generate_ticket(&mut self, service_type: ServiceType, is_vip: bool) -> Ticket {
        let counter_map = if is_vip { &mut self.vip_counters } else { &mut self.counters };
        let next_number = counter_map.entry(service_type.clone()).or_insert(0);
        *next_number += 1;
        
        let number = TicketNumber::new(service_type.clone(), *next_number, is_vip);
        let ticket = Ticket::new(number);
        let ticket_key = ticket.number.display();
        
        self.all_tickets.insert(ticket_key.clone(), ticket.clone());
        self.enqueue_ticket(ticket.clone());
        
        ticket
    }

    fn enqueue_ticket(&mut self, ticket: Ticket) {
        let service_type = ticket.number.service_type.clone();
        let queue = self.service_queues.entry(service_type).or_insert_with(VecDeque::new);
        
        if ticket.number.is_vip && !ticket.loses_vip_priority() {
            let insert_pos = queue.iter().position(|t| !t.number.is_vip).unwrap_or(queue.len());
            queue.insert(insert_pos, ticket);
        } else {
            queue.push_back(ticket);
        }
    }

    pub fn call_next(&mut self, window_id: u32) -> Result<Ticket, QueueError> {
        let window = self.windows.get_mut(&window_id)
            .ok_or(QueueError::WindowNotFound(window_id))?;
            
        if !window.is_active {
            return Err(QueueError::WindowNotFound(window_id));
        }
        
        if !window.is_free() {
            return Err(QueueError::WindowBusy(window_id));
        }
        
        let mut selected_ticket: Option<Ticket> = None;
        let mut selected_queue_type: Option<ServiceType> = None;
        
        for service_type in &window.service_types {
            if let Some(queue) = self.service_queues.get(service_type) {
                if let Some(ticket) = queue.front() {
                    if let Some(ref current) = selected_ticket {
                        let should_replace = if ticket.number.is_vip && !current.number.is_vip {
                            true
                        } else if ticket.number.is_vip == current.number.is_vip {
                            ticket.created_at < current.created_at
                        } else {
                            false
                        };
                        
                        if should_replace {
                            selected_ticket = Some(ticket.clone());
                            selected_queue_type = Some(service_type.clone());
                        }
                    } else {
                        selected_ticket = Some(ticket.clone());
                        selected_queue_type = Some(service_type.clone());
                    }
                }
            }
        }
        
        let ticket = selected_ticket.ok_or(QueueError::NoWaitingTickets)?;
        let queue_type = selected_queue_type.unwrap();
        
        if let Some(queue) = self.service_queues.get_mut(&queue_type) {
            queue.pop_front();
        }
        
        let mut ticket = ticket;
        ticket.status = TicketStatus::Called;
        ticket.call_count = 1;
        ticket.called_at = Some(std::time::Instant::now());
        
        let window = self.windows.get_mut(&window_id).unwrap();
        window.current_ticket = Some(ticket.clone());
        
        self.all_tickets.insert(ticket.number.display(), ticket.clone());
        
        self.call_history.push(CallRecord {
            ticket_number: ticket.number.clone(),
            window_id,
            call_count: 1,
            called_at: std::time::Instant::now(),
        });
        
        Ok(ticket)
    }

    pub fn recall_ticket(&mut self, window_id: u32) -> Result<Ticket, QueueError> {
        let window = self.windows.get_mut(&window_id)
            .ok_or(QueueError::WindowNotFound(window_id))?;
        
        let mut ticket = window.current_ticket.as_mut()
            .ok_or(QueueError::WindowNotFound(window_id))?
            .clone();
            
        if ticket.status != TicketStatus::Called {
            return Err(QueueError::InvalidTicketStatus(
                ticket.number.display(),
                "Ticket is not in called state".to_string()
            ));
        }
        
        if ticket.call_count >= 3 {
            return self.handle_missed(window_id);
        }
        
        ticket.call_count += 1;
        
        self.call_history.push(CallRecord {
            ticket_number: ticket.number.clone(),
            window_id,
            call_count: ticket.call_count,
            called_at: std::time::Instant::now(),
        });
        
        let window = self.windows.get_mut(&window_id).unwrap();
        window.current_ticket = Some(ticket.clone());
        
        self.all_tickets.insert(ticket.number.display(), ticket.clone());
        
        Ok(ticket)
    }

    pub fn handle_missed(&mut self, window_id: u32) -> Result<Ticket, QueueError> {
        let window = self.windows.get_mut(&window_id)
            .ok_or(QueueError::WindowNotFound(window_id))?;
        
        let mut ticket = window.current_ticket.take()
            .ok_or(QueueError::WindowNotFound(window_id))?;
            
        ticket.status = TicketStatus::Missed;
        ticket.missed_count += 1;
        
        if !ticket.can_be_queued() {
            ticket.status = TicketStatus::Expired;
            self.all_tickets.insert(ticket.number.display(), ticket.clone());
            return Err(QueueError::TooManyMisses(ticket.number.display()));
        }
        
        self.enqueue_ticket(ticket.clone());
        self.all_tickets.insert(ticket.number.display(), ticket.clone());
        
        Ok(ticket)
    }

    pub fn mark_ticket_arrived(&mut self, window_id: u32) -> Result<Ticket, QueueError> {
        let window = self.windows.get_mut(&window_id)
            .ok_or(QueueError::WindowNotFound(window_id))?;
        
        let mut ticket = window.current_ticket.as_mut()
            .ok_or(QueueError::WindowNotFound(window_id))?
            .clone();
            
        if ticket.status != TicketStatus::Called {
            return Err(QueueError::InvalidTicketStatus(
                ticket.number.display(),
                "Ticket is not in called state".to_string()
            ));
        }
        
        ticket.status = TicketStatus::Processing;
        self.all_tickets.insert(ticket.number.display(), ticket.clone());
        
        let window = self.windows.get_mut(&window_id).unwrap();
        window.current_ticket = Some(ticket.clone());
        
        Ok(ticket)
    }

    pub fn complete_ticket(&mut self, window_id: u32) -> Result<Ticket, QueueError> {
        let window = self.windows.get_mut(&window_id)
            .ok_or(QueueError::WindowNotFound(window_id))?;
        
        let mut ticket = window.current_ticket.take()
            .ok_or(QueueError::WindowNotFound(window_id))?;
            
        ticket.status = TicketStatus::Completed;
        self.all_tickets.insert(ticket.number.display(), ticket.clone());
        
        Ok(ticket)
    }

    pub fn close_window(&mut self, window_id: u32) -> Result<(), QueueError> {
        let window = self.windows.get_mut(&window_id)
            .ok_or(QueueError::WindowNotFound(window_id))?;
            
        if !window.is_active {
            return Err(QueueError::WindowAlreadyClosed(window_id));
        }
        
        window.is_active = false;
        
        if let Some(mut ticket) = window.current_ticket.take() {
            ticket.status = TicketStatus::Waiting;
            self.enqueue_ticket(ticket);
        }
        
        Ok(())
    }

    pub fn open_window(&mut self, window_id: u32) -> Result<(), QueueError> {
        let window = self.windows.get_mut(&window_id)
            .ok_or(QueueError::WindowNotFound(window_id))?;
            
        if window.is_active {
            return Err(QueueError::WindowAlreadyOpen(window_id));
        }
        
        window.is_active = true;
        Ok(())
    }

    pub fn get_queue_length(&self, service_type: &ServiceType) -> usize {
        self.service_queues.get(service_type).map_or(0, |q| q.len())
    }

    pub fn get_ticket(&self, display: &str) -> Option<&Ticket> {
        self.all_tickets.get(display)
    }

    pub fn get_call_history(&self) -> &Vec<CallRecord> {
        &self.call_history
    }

    pub fn get_all_tickets(&self) -> Vec<&Ticket> {
        self.all_tickets.values().collect()
    }
}

impl Default for QueueManager {
    fn default() -> Self {
        Self::new()
    }
}
