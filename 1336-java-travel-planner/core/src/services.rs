use uuid::Uuid;
use crate::models::*;
use crate::errors::*;
use crate::utils::*;

const MIN_GAP_MINUTES: i64 = 10;

pub struct TravelPlannerService {
    data_store: DataStore,
}

impl TravelPlannerService {
    pub fn new() -> Self {
        Self {
            data_store: DataStore::default(),
        }
    }

    pub fn data_store(&self) -> &DataStore {
        &self.data_store
    }

    pub fn add_attraction(&mut self, new_attraction: NewAttraction) -> Attraction {
        let attraction = Attraction {
            id: Uuid::new_v4(),
            name: new_attraction.name,
            description: new_attraction.description,
            suggested_duration_minutes: new_attraction.suggested_duration_minutes,
            opening_time: new_attraction.opening_time,
            closing_time: new_attraction.closing_time,
        };
        self.data_store.attractions.insert(attraction.id, attraction.clone());
        attraction
    }

    pub fn get_all_attractions(&self) -> Vec<Attraction> {
        self.data_store.attractions.values().cloned().collect()
    }

    pub fn get_attraction(&self, id: Uuid) -> Option<Attraction> {
        self.data_store.attractions.get(&id).cloned()
    }

    pub fn add_transportation(&mut self, new_transportation: NewTransportation) -> Result<Transportation> {
        if !self.data_store.attractions.contains_key(&new_transportation.from_attraction_id) {
            return Err(TravelPlannerError::AttractionNotFound(new_transportation.from_attraction_id));
        }
        if !self.data_store.attractions.contains_key(&new_transportation.to_attraction_id) {
            return Err(TravelPlannerError::AttractionNotFound(new_transportation.to_attraction_id));
        }

        let transportation = Transportation {
            from_attraction_id: new_transportation.from_attraction_id,
            to_attraction_id: new_transportation.to_attraction_id,
            transport_type: new_transportation.transport_type,
            duration_minutes: new_transportation.duration_minutes,
        };
        self.data_store.transportations.push(transportation.clone());
        Ok(transportation)
    }

    pub fn get_all_transportations(&self) -> Vec<Transportation> {
        self.data_store.transportations.clone()
    }

    pub fn create_itinerary(&mut self, new_itinerary: NewItinerary) -> Itinerary {
        let itinerary = Itinerary {
            id: Uuid::new_v4(),
            name: new_itinerary.name,
            nodes: Vec::new(),
        };
        self.data_store.itineraries.insert(itinerary.id, itinerary.clone());
        itinerary
    }

    pub fn get_all_itineraries(&self) -> Vec<Itinerary> {
        self.data_store.itineraries.values().cloned().collect()
    }

    pub fn get_itinerary(&self, id: Uuid) -> Option<Itinerary> {
        self.data_store.itineraries.get(&id).cloned()
    }

    pub fn add_nodes_to_itinerary(&mut self, itinerary_id: Uuid, new_nodes: Vec<NewItineraryNode>) -> Result<Itinerary> {
        let itinerary = self.data_store.itineraries.get(&itinerary_id)
            .ok_or(TravelPlannerError::ItineraryNotFound(itinerary_id))?
            .clone();

        let existing_nodes = itinerary.nodes.clone();

        let new_nodes_with_ids: Vec<ItineraryNode> = new_nodes
            .into_iter()
            .map(|n| ItineraryNode {
                id: Uuid::new_v4(),
                attraction_id: n.attraction_id,
                arrival_time: n.arrival_time,
                departure_time: n.departure_time,
            })
            .collect();

        self.validate_nodes(&existing_nodes, &new_nodes_with_ids)?;

        let itinerary = self.data_store.itineraries.get_mut(&itinerary_id).unwrap();
        itinerary.nodes.extend(new_nodes_with_ids);
        itinerary.nodes.sort_by(|a, b| a.arrival_time.cmp(&b.arrival_time));

        Ok(itinerary.clone())
    }

    fn validate_nodes(&self, existing_nodes: &[ItineraryNode], new_nodes: &[ItineraryNode]) -> Result<()> {
        let mut errors: Vec<TravelPlannerError> = Vec::new();

        for node in new_nodes {
            if let Err(e) = self.validate_single_node(node) {
                errors.extend(match e {
                    TravelPlannerError::MultipleErrors(es) => es,
                    _ => vec![e],
                });
            }
        }

        let all_nodes: Vec<ItineraryNode> = existing_nodes
            .iter()
            .chain(new_nodes.iter())
            .cloned()
            .collect();

        if let Err(e) = self.validate_no_overlap(&all_nodes) {
            errors.extend(match e {
                TravelPlannerError::MultipleErrors(es) => es,
                _ => vec![e],
            });
        }

        if let Err(e) = self.validate_same_day(&all_nodes) {
            errors.extend(match e {
                TravelPlannerError::MultipleErrors(es) => es,
                _ => vec![e],
            });
        }

        if let Err(e) = self.validate_transportation_times(&all_nodes) {
            errors.extend(match e {
                TravelPlannerError::MultipleErrors(es) => es,
                _ => vec![e],
            });
        }

        if errors.is_empty() {
            Ok(())
        } else {
            Err(TravelPlannerError::MultipleErrors(errors))
        }
    }

    fn validate_single_node(&self, node: &ItineraryNode) -> Result<()> {
        let mut errors: Vec<TravelPlannerError> = Vec::new();

        if node.departure_time <= node.arrival_time {
            errors.push(TravelPlannerError::DepartureBeforeArrival { node_id: node.id });
        }

        let arrival_date = node.arrival_time.naive_utc().date();
        let departure_date = node.departure_time.naive_utc().date();
        if arrival_date != departure_date {
            errors.push(TravelPlannerError::CrossesMidnight {
                node_id: node.id,
                arrival: node.arrival_time,
                departure: node.departure_time,
            });
        }

        let attraction = self.data_store.attractions.get(&node.attraction_id)
            .ok_or(TravelPlannerError::AttractionNotFound(node.attraction_id))?;

        let arrival_time = datetime_to_naive_time(node.arrival_time);
        let departure_time = datetime_to_naive_time(node.departure_time);

        let arrival_ok = is_time_in_range(arrival_time, attraction.opening_time, attraction.closing_time);
        let departure_ok = is_time_in_range(departure_time, attraction.opening_time, attraction.closing_time);

        if !arrival_ok || !departure_ok {
            errors.push(TravelPlannerError::AttractionClosed {
                node_id: node.id,
                attraction_name: attraction.name.clone(),
                open: attraction.opening_time.format("%H:%M").to_string(),
                close: attraction.closing_time.format("%H:%M").to_string(),
            });
        }

        if errors.is_empty() {
            Ok(())
        } else {
            Err(TravelPlannerError::MultipleErrors(errors))
        }
    }

    fn validate_no_overlap(&self, all_nodes: &[ItineraryNode]) -> Result<()> {
        let mut errors: Vec<TravelPlannerError> = Vec::new();
        let mut sorted_nodes: Vec<ItineraryNode> = all_nodes.to_vec();
        sorted_nodes.sort_by(|a, b| a.arrival_time.cmp(&b.arrival_time));

        for i in 0..sorted_nodes.len() {
            for j in (i + 1)..sorted_nodes.len() {
                let node1 = &sorted_nodes[i];
                let node2 = &sorted_nodes[j];

                if node1.departure_time > node2.arrival_time {
                    errors.push(TravelPlannerError::NodeOverlap {
                        node1: node1.id,
                        node2: node2.id,
                    });
                } else if node1.departure_time == node2.arrival_time {
                    continue;
                } else {
                    let gap = duration_seconds(node1.departure_time, node2.arrival_time);
                    if gap < MIN_GAP_MINUTES * 60 {
                        errors.push(TravelPlannerError::TimeConflict {
                            node1: node1.id,
                            node2: node2.id,
                            gap_seconds: gap,
                        });
                    }
                }
            }
        }

        if errors.is_empty() {
            Ok(())
        } else {
            Err(TravelPlannerError::MultipleErrors(errors))
        }
    }

    fn validate_same_day(&self, _all_nodes: &[ItineraryNode]) -> Result<()> {
        Ok(())
    }

    fn validate_transportation_times(&self, all_nodes: &[ItineraryNode]) -> Result<()> {
        let mut errors: Vec<TravelPlannerError> = Vec::new();
        let mut sorted_nodes: Vec<ItineraryNode> = all_nodes.to_vec();
        sorted_nodes.sort_by(|a, b| a.arrival_time.cmp(&b.arrival_time));

        for i in 1..sorted_nodes.len() {
            let prev = &sorted_nodes[i - 1];
            let curr = &sorted_nodes[i];

            let prev_date = prev.arrival_time.naive_utc().date();
            let curr_date = curr.arrival_time.naive_utc().date();
            
            if prev_date == curr_date {
                let transport = self.find_transportation(prev.attraction_id, curr.attraction_id);

                match transport {
                    Some(t) => {
                        let available = duration_minutes(prev.departure_time, curr.arrival_time);
                        let required = t.duration_minutes as i64;

                        if available < required {
                            errors.push(TravelPlannerError::InsufficientTransportationTime {
                                prev_node: prev.id,
                                curr_node: curr.id,
                                required,
                                available,
                            });
                        }
                    }
                    None => {
                        errors.push(TravelPlannerError::TransportationNotFound {
                            from: prev.attraction_id,
                            to: curr.attraction_id,
                        });
                    }
                }
            }
        }

        if errors.is_empty() {
            Ok(())
        } else {
            Err(TravelPlannerError::MultipleErrors(errors))
        }
    }

    fn find_transportation(&self, from: Uuid, to: Uuid) -> Option<&Transportation> {
        self.data_store.transportations.iter().find(|t| {
            t.from_attraction_id == from && t.to_attraction_id == to
        })
    }
}

impl Default for TravelPlannerService {
    fn default() -> Self {
        Self::new()
    }
}
