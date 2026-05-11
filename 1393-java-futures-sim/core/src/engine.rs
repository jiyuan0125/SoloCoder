use crate::models::*;
use serde::Serialize;
use std::collections::HashMap;
use thiserror::Error;

#[derive(Debug, Error, Clone)]
pub enum TradingError {
    #[error("合约不存在: {0}")]
    ContractNotFound(String),
    #[error("用户不存在: {0}")]
    UserNotFound(String),
    #[error("可用资金不足: 需要 {required}, 可用 {available}")]
    InsufficientFunds { required: f64, available: f64 },
    #[error("持仓不足: 需要 {required} 手, 仅有 {available} 手")]
    InsufficientPosition { required: i32, available: i32 },
    #[error("手数必须为正")]
    InvalidLots,
}

#[derive(Debug, Clone, Serialize)]
pub struct TradeResult {
    pub user_id: String,
    pub action: String,
    pub contract_code: String,
    pub position_type: PositionType,
    pub lots: i32,
    pub price: f64,
    pub fee: f64,
    pub realized_pnl: f64,
}

pub struct TradingEngine {
    contracts: HashMap<String, Contract>,
    users: HashMap<String, User>,
    time: u64,
}

impl TradingEngine {
    pub fn new() -> Self {
        Self {
            contracts: HashMap::new(),
            users: HashMap::new(),
            time: 0,
        }
    }

    fn next_time(&mut self) -> u64 {
        self.time += 1;
        self.time
    }

    pub fn add_contract(&mut self, contract: Contract) {
        self.contracts.insert(contract.code.clone(), contract);
    }

    pub fn get_contract(&self, code: &str) -> Option<&Contract> {
        self.contracts.get(code)
    }

    pub fn get_contracts(&self) -> Vec<&Contract> {
        self.contracts.values().collect()
    }

    pub fn register_user(&mut self, user_id: &str) -> &User {
        self.users
            .entry(user_id.to_string())
            .or_insert_with(|| User::new(user_id, INITIAL_FUNDS))
    }

    pub fn get_user(&self, user_id: &str) -> Option<&User> {
        self.users.get(user_id)
    }

    pub fn get_users(&self) -> Vec<&User> {
        self.users.values().collect()
    }

    fn get_contract_mut(&mut self, code: &str) -> Option<&mut Contract> {
        self.contracts.get_mut(code)
    }

    fn get_user_mut(&mut self, user_id: &str) -> Option<&mut User> {
        self.users.get_mut(user_id)
    }

    pub fn open_position(
        &mut self,
        user_id: &str,
        contract_code: &str,
        position_type: PositionType,
        lots: i32,
    ) -> Result<TradeResult, TradingError> {
        if lots <= 0 {
            return Err(TradingError::InvalidLots);
        }

        let contract = self
            .get_contract(contract_code)
            .ok_or_else(|| TradingError::ContractNotFound(contract_code.to_string()))?
            .clone();

        let user_exists = self.get_user(user_id).is_some();
        if !user_exists {
            return Err(TradingError::UserNotFound(user_id.to_string()));
        }

        let margin_needed = contract.required_margin(contract.current_price, lots);
        let contract_value = contract.contract_value(contract.current_price, lots);
        let fee = contract_value * FEE_RATE;
        let total_required = margin_needed + fee;

        let available = {
            let user = self
                .get_user(user_id)
                .ok_or_else(|| TradingError::UserNotFound(user_id.to_string()))?;
            user.available_funds()
        };

        if available < total_required {
            return Err(TradingError::InsufficientFunds {
                required: total_required,
                available,
            });
        }

        let open_time = self.next_time();

        {
            let user = self
                .get_user_mut(user_id)
                .ok_or_else(|| TradingError::UserNotFound(user_id.to_string()))?;
            user.total_funds -= fee;
            user.used_margin += margin_needed;
            user.realized_pnl -= fee;

            let position = Position {
                contract_code: contract_code.to_string(),
                position_type,
                lots,
                open_price: contract.current_price,
                open_time,
            };
            user.add_position(position);
        }

        Ok(TradeResult {
            user_id: user_id.to_string(),
            action: "OPEN".to_string(),
            contract_code: contract_code.to_string(),
            position_type,
            lots,
            price: contract.current_price,
            fee,
            realized_pnl: 0.0,
        })
    }

    pub fn close_position(
        &mut self,
        user_id: &str,
        contract_code: &str,
        position_type: PositionType,
        lots: i32,
    ) -> Result<Vec<TradeResult>, TradingError> {
        if lots <= 0 {
            return Err(TradingError::InvalidLots);
        }

        let contract = self
            .get_contract(contract_code)
            .ok_or_else(|| TradingError::ContractNotFound(contract_code.to_string()))?
            .clone();

        let user = self
            .get_user_mut(user_id)
            .ok_or_else(|| TradingError::UserNotFound(user_id.to_string()))?;

        let available_lots = user.total_lots(contract_code, position_type);
        if available_lots < lots {
            return Err(TradingError::InsufficientPosition {
                required: lots,
                available: available_lots,
            });
        }

        let mut remaining = lots;
        let mut results = Vec::new();
        let mut position_ids_to_remove = Vec::new();
        let mut positions_to_update = Vec::new();
        let mut total_fee = 0.0;
        let mut total_pnl = 0.0;
        let mut margin_freed = 0.0;

        for (pos_id, pos) in user.positions.iter() {
            if remaining <= 0 {
                break;
            }
            if pos.contract_code != contract_code || pos.position_type != position_type {
                continue;
            }

            let lots_to_close = pos.lots.min(remaining);
            let contract_value = contract.contract_value(contract.current_price, lots_to_close);
            let fee = contract_value * FEE_RATE;

            let diff = match position_type {
                PositionType::Long => contract.current_price - pos.open_price,
                PositionType::Short => pos.open_price - contract.current_price,
            };
            let pnl = diff * contract.contract_multiplier * lots_to_close as f64;

            total_fee += fee;
            total_pnl += pnl;

            let margin_for_lots = contract.required_margin(pos.open_price, lots_to_close);
            margin_freed += margin_for_lots;

            results.push(TradeResult {
                user_id: user_id.to_string(),
                action: "CLOSE".to_string(),
                contract_code: contract_code.to_string(),
                position_type,
                lots: lots_to_close,
                price: contract.current_price,
                fee,
                realized_pnl: pnl,
            });

            if lots_to_close == pos.lots {
                position_ids_to_remove.push(*pos_id);
            } else {
                positions_to_update.push((*pos_id, lots_to_close));
            }

            remaining -= lots_to_close;
        }

        for id in position_ids_to_remove {
            user.positions.remove(&id);
        }
        for (id, closed) in positions_to_update {
            if let Some(pos) = user.positions.get_mut(&id) {
                pos.lots -= closed;
            }
        }

        user.total_funds += total_pnl;
        user.total_funds -= total_fee;
        user.used_margin -= margin_freed;
        user.realized_pnl += total_pnl - total_fee;

        Ok(results)
    }

    fn check_force_liquidation(&mut self, user_id: &str) -> Vec<TradeResult> {
        let mut all_results = Vec::new();
        let contracts_to_check: Vec<(String, f64)> = {
            let user = match self.get_user(user_id) {
                Some(u) => u,
                None => return all_results,
            };

            let contracts: Vec<_> = self.contracts.values().cloned().collect();
            let effective_available =
                user.total_funds - user.used_margin + user.all_unrealized_pnl(contracts.iter());

            if effective_available >= 0.0 {
                return all_results;
            }

            user.positions
                .values()
                .map(|p| (p.contract_code.clone(), p.position_type as u8))
                .collect::<std::collections::HashSet<_>>()
                .into_iter()
                .map(|(c, _)| c)
                .zip(std::iter::repeat(0.0))
                .collect()
        };

        let contract_codes: Vec<String> = contracts_to_check.iter().map(|(c, _)| c.clone()).collect();
        for contract_code in contract_codes {
            for &pt in &[PositionType::Long, PositionType::Short] {
                loop {
                    let lots = match self.get_user(user_id) {
                        Some(u) => u.total_lots(&contract_code, pt),
                        None => break,
                    };
                    if lots <= 0 {
                        break;
                    }

                    let close_results =
                        match self.close_position(user_id, &contract_code, pt, lots) {
                            Ok(r) => r,
                            Err(_) => break,
                        };
                    all_results.extend(close_results);

                    let user = match self.get_user(user_id) {
                        Some(u) => u.clone(),
                        None => break,
                    };
                    let contracts: Vec<_> = self.contracts.values().cloned().collect();
                    let effective_available = user.total_funds - user.used_margin
                        + user.all_unrealized_pnl(contracts.iter());
                    if effective_available >= 0.0 {
                        break;
                    }
                }
            }
        }

        all_results
    }

    pub fn update_price(
        &mut self,
        contract_code: &str,
        new_price: f64,
    ) -> Result<Vec<TradeResult>, TradingError> {
        {
            let contract = self
                .get_contract_mut(contract_code)
                .ok_or_else(|| TradingError::ContractNotFound(contract_code.to_string()))?;
            contract.current_price = new_price;
        }

        let user_ids: Vec<String> = self.users.keys().cloned().collect();
        let mut all_liquidations = Vec::new();

        for user_id in user_ids {
            let has_position = self
                .get_user(&user_id)
                .map(|u| {
                    u.positions
                        .values()
                        .any(|p| p.contract_code == contract_code)
                })
                .unwrap_or(false);

            if has_position {
                let results = self.check_force_liquidation(&user_id);
                all_liquidations.extend(results);
            }
        }

        Ok(all_liquidations)
    }

    pub fn settle(&mut self) -> Vec<TradeResult> {
        let mut all_results = Vec::new();
        let contracts: Vec<Contract> = self.contracts.values().cloned().collect();
        let user_ids: Vec<String> = self.users.keys().cloned().collect();

        for contract in &contracts {
            if let Some(c) = self.get_contract_mut(&contract.code) {
                c.settlement_price = c.current_price;
            }
        }

        for user_id in user_ids {
            let settle_pnl = self.settle_user_positions(&user_id);
            all_results.extend(settle_pnl);
        }

        for user_id in self.users.keys().cloned().collect::<Vec<_>>() {
            let results = self.check_force_liquidation(&user_id);
            all_results.extend(results);
        }

        all_results
    }

    fn settle_user_positions(&mut self, user_id: &str) -> Vec<TradeResult> {
        let mut results = Vec::new();
        let contracts: Vec<Contract> = self.contracts.values().cloned().collect();
        let user_exists = self.get_user(user_id).is_some();
        if !user_exists {
            return results;
        }

        let open_time = self.next_time();

        {
            let user = match self.get_user_mut(user_id) {
                Some(u) => u,
                None => return results,
            };

            let unrealized = user.all_unrealized_pnl(contracts.iter());
            if unrealized != 0.0 {
                user.total_funds += unrealized;
                user.realized_pnl += unrealized;
            }

            user.used_margin = 0.0;
            for pos in user.positions.values() {
                if let Some(contract) = contracts.iter().find(|c| c.code == pos.contract_code) {
                    let new_margin = contract.required_margin(contract.current_price, pos.lots);
                    user.used_margin += new_margin;
                }
            }

            let old_positions: Vec<Position> = user.positions.values().cloned().collect();
            user.positions.clear();

            for pos in old_positions {
                if let Some(contract) = contracts.iter().find(|c| c.code == pos.contract_code) {
                    results.push(TradeResult {
                        user_id: user_id.to_string(),
                        action: "SETTLE".to_string(),
                        contract_code: pos.contract_code.clone(),
                        position_type: pos.position_type,
                        lots: pos.lots,
                        price: contract.current_price,
                        fee: 0.0,
                        realized_pnl: 0.0,
                    });

                    let new_pos = Position {
                        contract_code: pos.contract_code,
                        position_type: pos.position_type,
                        lots: pos.lots,
                        open_price: contract.current_price,
                        open_time,
                    };
                    user.add_position(new_pos);
                }
            }
        }

        results
    }
}

impl Default for TradingEngine {
    fn default() -> Self {
        Self::new()
    }
}
