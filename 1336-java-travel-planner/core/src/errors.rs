use thiserror::Error;
use chrono::{DateTime, Utc};
use uuid::Uuid;

#[derive(Debug, Error)]
pub enum TravelPlannerError {
    #[error("Attraction not found: {0}")]
    AttractionNotFound(Uuid),
    
    #[error("Itinerary not found: {0}")]
    ItineraryNotFound(Uuid),
    
    #[error("No transportation found from {from} to {to}")]
    TransportationNotFound {
        from: Uuid,
        to: Uuid,
    },
    
    #[error("Node {node_id}: Departure time must be after arrival time")]
    DepartureBeforeArrival {
        node_id: Uuid,
    },
    
    #[error("Node {node_id}: Time range crosses midnight (arrival: {arrival}, departure: {departure})")]
    CrossesMidnight {
        node_id: Uuid,
        arrival: DateTime<Utc>,
        departure: DateTime<Utc>,
    },
    
    #[error("Node {node_id}: Attraction {attraction_name} is not open during scheduled time. Opens at {open}, closes at {close}")]
    AttractionClosed {
        node_id: Uuid,
        attraction_name: String,
        open: String,
        close: String,
    },
    
    #[error("Insufficient transportation time between node {prev_node} and {curr_node}. Need at least {required} minutes, but only have {available} minutes")]
    InsufficientTransportationTime {
        prev_node: Uuid,
        curr_node: Uuid,
        required: i64,
        available: i64,
    },
    
    #[error("Time conflict between node {node1} and {node2}. Gap of {gap_seconds} seconds is less than 10 minutes")]
    TimeConflict {
        node1: Uuid,
        node2: Uuid,
        gap_seconds: i64,
    },
    
    #[error("Nodes overlap between {node1} and {node2}")]
    NodeOverlap {
        node1: Uuid,
        node2: Uuid,
    },
    
    #[error("Multiple validation errors: {0:?}")]
    MultipleErrors(Vec<TravelPlannerError>),
}

pub type Result<T> = std::result::Result<T, TravelPlannerError>;
