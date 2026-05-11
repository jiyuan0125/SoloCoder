use serde::{Deserialize, Serialize};
use std::collections::BTreeMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Contract {
    pub code: String,
    pub current_price: f64,
    pub contract_multiplier: f64,
    pub margin_ratio: f64,
    pub settlement_price: f64,
}

impl Contract {
    pub fn new(code: &str, price: f64, multiplier: f64, margin_ratio: f64) -> Self {
        Self {
            code: code.to_string(),
            current_price: price,
            contract_multiplier: multiplier,
            margin_ratio,
            settlement_price: price,
        }
    }

    pub fn contract_value(&self, price: f64, lots: i32) -> f64 {
        price * self.contract_multiplier * lots as f64
    }

    pub fn required_margin(&self, price: f64, lots: i32) -> f64 {
        self.contract_value(price, lots) * self.margin_ratio
    }
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum PositionType {
    Long,
    Short,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Position {
    pub contract_code: String,
    pub position_type: PositionType,
    pub lots: i32,
    pub open_price: f64,
    pub open_time: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct User {
    pub id: String,
    pub total_funds: f64,
    pub used_margin: f64,
    pub realized_pnl: f64,
    pub positions: BTreeMap<u64, Position>,
    next_position_id: u64,
}

impl User {
    pub fn new(id: &str, initial_funds: f64) -> Self {
        Self {
            id: id.to_string(),
            total_funds: initial_funds,
            used_margin: 0.0,
            realized_pnl: 0.0,
            positions: BTreeMap::new(),
            next_position_id: 1,
        }
    }

    pub fn available_funds(&self) -> f64 {
        self.total_funds - self.used_margin
    }

    pub fn add_position(&mut self, pos: Position) -> u64 {
        let id = self.next_position_id;
        self.next_position_id += 1;
        self.positions.insert(id, pos);
        id
    }

    pub fn remove_position(&mut self, id: u64) -> Option<Position> {
        self.positions.remove(&id)
    }

    pub fn total_lots(&self, contract_code: &str, position_type: PositionType) -> i32 {
        self.positions
            .values()
            .filter(|p| p.contract_code == contract_code && p.position_type == position_type)
            .map(|p| p.lots)
            .sum()
    }

    pub fn unrealized_pnl(&self, contract: &Contract) -> f64 {
        let mut pnl = 0.0;
        for pos in self.positions.values() {
            if pos.contract_code == contract.code {
                pnl += position_unrealized_pnl(pos, contract);
            }
        }
        pnl
    }

    pub fn all_unrealized_pnl<'a, I>(&'a self, contracts: I) -> f64
    where
        I: Iterator<Item = &'a Contract>,
    {
        let mut pnl = 0.0;
        for contract in contracts {
            pnl += self.unrealized_pnl(contract);
        }
        pnl
    }
}

pub fn position_unrealized_pnl(pos: &Position, contract: &Contract) -> f64 {
    let diff = match pos.position_type {
        PositionType::Long => contract.current_price - pos.open_price,
        PositionType::Short => pos.open_price - contract.current_price,
    };
    diff * contract.contract_multiplier * pos.lots as f64
}

pub const FEE_RATE: f64 = 0.0001;
pub const INITIAL_FUNDS: f64 = 1_000_000.0;
