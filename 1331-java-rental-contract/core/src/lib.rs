pub mod models;
pub mod storage;
pub mod service;
pub mod error;

pub use error::RentalError;
pub use models::*;
pub use service::RentalService;
pub use storage::InMemoryStorage;
