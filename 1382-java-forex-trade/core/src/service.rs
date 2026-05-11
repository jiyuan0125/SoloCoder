use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use chrono::Utc;

use crate::error::{ForexError, Result};
use crate::models::{Account, Balance, Currency, CurrencyPair, ExchangeRate, Order, OrderSide, OrderStatus};

pub fn round_amount(amount: f64, currency: Currency) -> f64 {
    let precision = currency.precision();
    let multiplier = 10f64.powi(precision as i32);
    (amount * multiplier).round() / multiplier
}

pub struct ForexSystem {
    accounts: Arc<Mutex<HashMap<String, Account>>>,
    balances: Arc<Mutex<HashMap<(String, Currency), Balance>>>,
    orders: Arc<Mutex<HashMap<String, Order>>>,
    rates: Arc<Mutex<HashMap<CurrencyPair, ExchangeRate>>>,
}

impl ForexSystem {
    pub fn new() -> Self {
        let mut rates = HashMap::new();
        rates.insert(
            CurrencyPair::new(Currency::USD, Currency::CNY),
            ExchangeRate::new(CurrencyPair::new(Currency::USD, Currency::CNY), 7.2450, 7.2550),
        );
        rates.insert(
            CurrencyPair::new(Currency::EUR, Currency::CNY),
            ExchangeRate::new(CurrencyPair::new(Currency::EUR, Currency::CNY), 7.8500, 7.8600),
        );
        rates.insert(
            CurrencyPair::new(Currency::USD, Currency::JPY),
            ExchangeRate::new(CurrencyPair::new(Currency::USD, Currency::JPY), 149.50, 149.60),
        );
        rates.insert(
            CurrencyPair::new(Currency::EUR, Currency::USD),
            ExchangeRate::new(CurrencyPair::new(Currency::EUR, Currency::USD), 1.0830, 1.0840),
        );

        Self {
            accounts: Arc::new(Mutex::new(HashMap::new())),
            balances: Arc::new(Mutex::new(HashMap::new())),
            orders: Arc::new(Mutex::new(HashMap::new())),
            rates: Arc::new(Mutex::new(rates)),
        }
    }

    pub fn create_account(&self, name: String, is_reviewer: bool) -> Account {
        let account = Account::new(name, is_reviewer);
        let id = account.id.clone();
        
        let mut accounts = self.accounts.lock().unwrap();
        accounts.insert(id.clone(), account.clone());

        let mut balances = self.balances.lock().unwrap();
        let initial_balance = if is_reviewer { 0.0 } else { 100000.0 };
        balances.insert((id.clone(), Currency::CNY), Balance::new(initial_balance));
        balances.insert((id.clone(), Currency::USD), Balance::new(0.0));
        balances.insert((id.clone(), Currency::EUR), Balance::new(0.0));
        balances.insert((id.clone(), Currency::JPY), Balance::new(0.0));

        account
    }

    pub fn get_account(&self, account_id: &str) -> Result<Account> {
        let accounts = self.accounts.lock().unwrap();
        accounts.get(account_id).cloned().ok_or(ForexError::AccountNotFound(account_id.to_string()))
    }

    pub fn get_balance(&self, account_id: &str, currency: Currency) -> Result<Balance> {
        let balances = self.balances.lock().unwrap();
        balances
            .get(&(account_id.to_string(), currency))
            .cloned()
            .ok_or(ForexError::AccountNotFound(account_id.to_string()))
    }

    pub fn get_all_balances(&self, account_id: &str) -> Result<HashMap<Currency, Balance>> {
        self.get_account(account_id)?;
        
        let balances = self.balances.lock().unwrap();
        let mut result = HashMap::new();
        
        for ((id, currency), balance) in balances.iter() {
            if id == account_id {
                result.insert(*currency, balance.clone());
            }
        }
        
        Ok(result)
    }

    pub fn get_all_rates(&self) -> Vec<ExchangeRate> {
        let rates = self.rates.lock().unwrap();
        rates.values().cloned().collect()
    }

    pub fn get_rate(&self, pair: CurrencyPair) -> Result<ExchangeRate> {
        let rates = self.rates.lock().unwrap();
        rates.get(&pair).cloned().ok_or(ForexError::CurrencyPairNotFound(pair.to_string()))
    }

    fn calculate_usd_equivalent(&self, amount: f64, currency: Currency) -> f64 {
        if currency == Currency::USD {
            return amount;
        }

        let rates = self.rates.lock().unwrap();
        
        if currency == Currency::CNY {
            if let Some(rate) = rates.get(&CurrencyPair::new(Currency::USD, Currency::CNY)) {
                return amount / ((rate.bid + rate.ask) / 2.0);
            }
        }
        
        if currency == Currency::EUR {
            if let Some(rate) = rates.get(&CurrencyPair::new(Currency::EUR, Currency::USD)) {
                return amount * ((rate.bid + rate.ask) / 2.0);
            }
        }
        
        if currency == Currency::JPY {
            if let Some(rate) = rates.get(&CurrencyPair::new(Currency::USD, Currency::JPY)) {
                return amount / ((rate.bid + rate.ask) / 2.0);
            }
        }

        amount
    }

    pub fn place_order(
        &self,
        account_id: &str,
        pair: CurrencyPair,
        side: OrderSide,
        amount: f64,
    ) -> Result<Order> {
        if amount <= 0.0 {
            return Err(ForexError::InvalidAmount);
        }

        self.get_account(account_id)?;

        let rate = self.get_rate(pair)?;
        let amount = round_amount(amount, pair.base);
        
        let (rate_value, frozen_currency, frozen_amount) = match side {
            OrderSide::Buy => {
                let frozen = round_amount(amount * rate.ask, pair.quote);
                (rate.ask, pair.quote, frozen)
            }
            OrderSide::Sell => {
                let frozen = round_amount(amount, pair.base);
                (rate.bid, pair.base, frozen)
            }
        };

        let usd_equivalent = self.calculate_usd_equivalent(
            match side {
                OrderSide::Buy => frozen_amount,
                OrderSide::Sell => frozen_amount,
            },
            frozen_currency,
        );

        {
            let mut balances = self.balances.lock().unwrap();
            let balance = balances
                .get_mut(&(account_id.to_string(), frozen_currency))
                .ok_or(ForexError::AccountNotFound(account_id.to_string()))?;

            if balance.available < frozen_amount {
                return Err(ForexError::InsufficientBalance);
            }

            balance.available -= frozen_amount;
            balance.frozen += frozen_amount;
        }

        let order = Order::new(
            account_id.to_string(),
            pair,
            side,
            amount,
            rate_value,
            usd_equivalent,
        );

        let order_id = order.id.clone();
        let mut orders = self.orders.lock().unwrap();
        orders.insert(order_id, order.clone());

        Ok(order)
    }

    pub fn approve_order(&self, reviewer_id: &str, order_id: &str, comment: Option<String>) -> Result<Order> {
        let reviewer = self.get_account(reviewer_id)?;
        if !reviewer.is_reviewer {
            return Err(ForexError::InvalidOrderStatus);
        }

        let mut orders = self.orders.lock().unwrap();
        let order = orders
            .get_mut(order_id)
            .ok_or(ForexError::OrderNotFound(order_id.to_string()))?;

        if order.status != OrderStatus::PendingApproval {
            return Err(ForexError::InvalidOrderStatus);
        }

        order.status = OrderStatus::Approved;
        order.approval_comment = comment;

        Ok(order.clone())
    }

    pub fn reject_order(&self, reviewer_id: &str, order_id: &str, comment: Option<String>) -> Result<Order> {
        let reviewer = self.get_account(reviewer_id)?;
        if !reviewer.is_reviewer {
            return Err(ForexError::InvalidOrderStatus);
        }

        let (order, frozen_currency, frozen_amount) = {
            let mut orders = self.orders.lock().unwrap();
            let order = orders
                .get_mut(order_id)
                .ok_or(ForexError::OrderNotFound(order_id.to_string()))?;

            if order.status != OrderStatus::PendingApproval {
                return Err(ForexError::InvalidOrderStatus);
            }

            let (frozen_currency, frozen_amount) = match order.side {
                OrderSide::Buy => (order.pair.quote, round_amount(order.amount * order.rate, order.pair.quote)),
                OrderSide::Sell => (order.pair.base, round_amount(order.amount, order.pair.base)),
            };

            order.status = OrderStatus::Rejected;
            order.approval_comment = comment;

            (order.clone(), frozen_currency, frozen_amount)
        };

        let mut balances = self.balances.lock().unwrap();
        let balance = balances
            .get_mut(&(order.account_id.clone(), frozen_currency))
            .ok_or(ForexError::AccountNotFound(order.account_id.clone()))?;

        balance.frozen -= frozen_amount;
        balance.available += frozen_amount;

        Ok(order)
    }

    pub fn settle_order(&self, order_id: &str) -> Result<Order> {
        let (order, account_id, side, pair, amount, rate) = {
            let mut orders = self.orders.lock().unwrap();
            let order = orders
                .get_mut(order_id)
                .ok_or(ForexError::OrderNotFound(order_id.to_string()))?;

            if order.status != OrderStatus::Pending && order.status != OrderStatus::Approved {
                return Err(ForexError::InvalidOrderStatus);
            }

            order.status = OrderStatus::Settled;

            (
                order.clone(),
                order.account_id.clone(),
                order.side,
                order.pair,
                order.amount,
                order.rate,
            )
        };

        let mut balances = self.balances.lock().unwrap();
        
        match side {
            OrderSide::Buy => {
                let counter_amount = round_amount(amount * rate, pair.quote);
                let receive_amount = round_amount(amount, pair.base);
                
                let quote_balance = balances
                    .get_mut(&(account_id.clone(), pair.quote))
                    .ok_or(ForexError::AccountNotFound(account_id.clone()))?;
                quote_balance.frozen -= counter_amount;
                
                let base_balance = balances
                    .get_mut(&(account_id.clone(), pair.base))
                    .ok_or(ForexError::AccountNotFound(account_id.clone()))?;
                base_balance.available += receive_amount;
            }
            OrderSide::Sell => {
                let counter_amount = round_amount(amount * rate, pair.quote);
                let frozen_amount = round_amount(amount, pair.base);
                
                let base_balance = balances
                    .get_mut(&(account_id.clone(), pair.base))
                    .ok_or(ForexError::AccountNotFound(account_id.clone()))?;
                base_balance.frozen -= frozen_amount;
                
                let quote_balance = balances
                    .get_mut(&(account_id.clone(), pair.quote))
                    .ok_or(ForexError::AccountNotFound(account_id.clone()))?;
                quote_balance.available += counter_amount;
            }
        }

        Ok(order)
    }

    pub fn get_order(&self, order_id: &str) -> Result<Order> {
        let orders = self.orders.lock().unwrap();
        orders.get(order_id).cloned().ok_or(ForexError::OrderNotFound(order_id.to_string()))
    }

    pub fn get_orders_by_account(&self, account_id: &str) -> Result<Vec<Order>> {
        self.get_account(account_id)?;
        
        let orders = self.orders.lock().unwrap();
        let result: Vec<Order> = orders
            .values()
            .filter(|o| o.account_id == account_id)
            .cloned()
            .collect();
        
        Ok(result)
    }

    pub fn get_pending_approval_orders(&self) -> Vec<Order> {
        let orders = self.orders.lock().unwrap();
        orders
            .values()
            .filter(|o| o.status == OrderStatus::PendingApproval)
            .cloned()
            .collect()
    }

    pub fn get_ready_to_settle_orders(&self) -> Vec<Order> {
        let today = Utc::now().date_naive();
        let orders = self.orders.lock().unwrap();
        
        orders
            .values()
            .filter(|o| {
                (o.status == OrderStatus::Pending || o.status == OrderStatus::Approved)
                    && o.settlement_date <= today
            })
            .cloned()
            .collect()
    }

    pub fn process_settlement(&self) -> usize {
        let ready_orders = self.get_ready_to_settle_orders();
        let mut settled = 0;
        
        for order in ready_orders {
            if self.settle_order(&order.id).is_ok() {
                settled += 1;
            }
        }
        
        settled
    }
}

impl Default for ForexSystem {
    fn default() -> Self {
        Self::new()
    }
}
