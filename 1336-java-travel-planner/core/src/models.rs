use chrono::{DateTime, NaiveTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Attraction {
    pub id: Uuid,
    pub name: String,
    pub description: Option<String>,
    pub suggested_duration_minutes: u32,
    pub opening_time: NaiveTime,
    pub closing_time: NaiveTime,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
pub enum TransportationType {
    Walk,
    Bus,
    Taxi,
    Subway,
    Car,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Transportation {
    pub from_attraction_id: Uuid,
    pub to_attraction_id: Uuid,
    pub transport_type: TransportationType,
    pub duration_minutes: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ItineraryNode {
    pub id: Uuid,
    pub attraction_id: Uuid,
    pub arrival_time: DateTime<Utc>,
    pub departure_time: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Itinerary {
    pub id: Uuid,
    pub name: String,
    pub nodes: Vec<ItineraryNode>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AddItineraryNodesRequest {
    pub itinerary_id: Uuid,
    pub nodes: Vec<NewItineraryNode>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NewItineraryNode {
    pub attraction_id: Uuid,
    pub arrival_time: DateTime<Utc>,
    pub departure_time: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NewAttraction {
    pub name: String,
    pub description: Option<String>,
    pub suggested_duration_minutes: u32,
    pub opening_time: NaiveTime,
    pub closing_time: NaiveTime,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NewTransportation {
    pub from_attraction_id: Uuid,
    pub to_attraction_id: Uuid,
    pub transport_type: TransportationType,
    pub duration_minutes: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NewItinerary {
    pub name: String,
}

#[derive(Debug, Clone, Default)]
pub struct DataStore {
    pub attractions: HashMap<Uuid, Attraction>,
    pub transportations: Vec<Transportation>,
    pub itineraries: HashMap<Uuid, Itinerary>,
}
