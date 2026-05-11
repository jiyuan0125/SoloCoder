pub mod error;
pub mod models;
pub mod service;

pub use error::{ForexError, Result};
pub use models::{
    Account, Balance, Currency, CurrencyPair, ExchangeRate, Order, OrderSide, OrderStatus,
};
pub use service::{round_amount, ForexSystem};
