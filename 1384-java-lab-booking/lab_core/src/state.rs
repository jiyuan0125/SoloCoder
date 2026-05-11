use std::collections::HashMap;
use std::sync::{Arc, RwLock};

use uuid::Uuid;

use crate::models::{Booking, Consumable, Equipment, User};

#[derive(Debug, Default, Clone)]
pub struct AppState {
    pub inner: Arc<RwLock<AppStateInner>>,
}

#[derive(Debug, Default)]
pub struct AppStateInner {
    pub users: HashMap<Uuid, User>,
    pub equipment: HashMap<Uuid, Equipment>,
    pub consumables: HashMap<Uuid, Consumable>,
    pub bookings: HashMap<Uuid, Booking>,
}

impl AppState {
    pub fn new() -> Self {
        Self::default()
    }
}
