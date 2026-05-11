use chrono::NaiveDate;
use rust_decimal::Decimal;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use uuid::Uuid;

use crate::error::AppError;
use crate::models::*;

#[derive(Clone)]
pub struct AppState {
    inner: Arc<Mutex<AppStateInner>>,
}

struct AppStateInner {
    declarations: HashMap<Uuid, Declaration>,
    exchange_rates: HashMap<(NaiveDate, Currency), ExchangeRate>,
}

impl AppState {
    pub fn new() -> Self {
        Self {
            inner: Arc::new(Mutex::new(AppStateInner {
                declarations: HashMap::new(),
                exchange_rates: HashMap::new(),
            })),
        }
    }
}

impl Default for AppState {
    fn default() -> Self {
        Self::new()
    }
}

pub fn add_exchange_rate(state: &AppState, req: CreateExchangeRateRequest) -> Result<ExchangeRate, AppError> {
    let mut inner = state.inner.lock().map_err(|e| AppError::Internal(e.to_string()))?;
    let key = (req.date, req.currency);
    if inner.exchange_rates.contains_key(&key) {
        return Err(AppError::ExchangeRateAlreadyExists(
            format!("{:?}", req.currency),
            req.date.to_string(),
        ));
    }
    let rate = ExchangeRate {
        date: req.date,
        currency: req.currency,
        rate_to_cny: req.rate_to_cny,
    };
    inner.exchange_rates.insert(key, rate.clone());
    Ok(rate)
}

pub fn get_exchange_rate(state: &AppState, date: NaiveDate, currency: Currency) -> Result<Option<ExchangeRate>, AppError> {
    let inner = state.inner.lock().map_err(|e| AppError::Internal(e.to_string()))?;
    let key = (date, currency);
    Ok(inner.exchange_rates.get(&key).cloned())
}

pub fn list_exchange_rates(state: &AppState) -> Result<Vec<ExchangeRate>, AppError> {
    let inner = state.inner.lock().map_err(|e| AppError::Internal(e.to_string()))?;
    let mut rates: Vec<ExchangeRate> = inner.exchange_rates.values().cloned().collect();
    rates.sort_by(|a, b| b.date.cmp(&a.date).then_with(|| format!("{:?}", a.currency).cmp(&format!("{:?}", b.currency))));
    Ok(rates)
}

pub fn create_declaration(state: &AppState, req: CreateDeclarationRequest) -> Result<Declaration, AppError> {
    let mut inner = state.inner.lock().map_err(|e| AppError::Internal(e.to_string()))?;
    for existing in inner.declarations.values() {
        if existing.declaration_no == req.declaration_no {
            return Err(AppError::DeclarationNoAlreadyExists(req.declaration_no));
        }
    }
    let decl = Declaration {
        id: Uuid::new_v4(),
        declaration_no: req.declaration_no,
        items: req.items,
        status: DeclarationStatus::Draft,
        declare_date: None,
        total_cny: None,
    };
    inner.declarations.insert(decl.id, decl.clone());
    Ok(decl)
}

pub fn get_declaration(state: &AppState, id: Uuid) -> Result<Option<Declaration>, AppError> {
    let inner = state.inner.lock().map_err(|e| AppError::Internal(e.to_string()))?;
    Ok(inner.declarations.get(&id).cloned())
}

pub fn list_declarations(state: &AppState) -> Result<Vec<DeclarationSummary>, AppError> {
    let inner = state.inner.lock().map_err(|e| AppError::Internal(e.to_string()))?;
    let mut list: Vec<DeclarationSummary> = inner.declarations.values().map(|d| d.summary()).collect();
    list.sort_by(|a, b| b.declaration_no.cmp(&a.declaration_no));
    Ok(list)
}

pub fn update_declaration(
    state: &AppState,
    id: Uuid,
    req: UpdateDeclarationRequest,
) -> Result<Declaration, AppError> {
    let mut inner = state.inner.lock().map_err(|e| AppError::Internal(e.to_string()))?;
    
    {
        let decl = inner
            .declarations
            .get(&id)
            .ok_or_else(|| AppError::DeclarationNotFound(id.to_string()))?;

        if decl.status != DeclarationStatus::Draft {
            return Err(AppError::InvalidStatus {
                expected: "Draft".to_string(),
                actual: format!("{:?}", decl.status),
            });
        }

        if let Some(ref new_no) = req.declaration_no {
            if new_no != &decl.declaration_no {
                for existing in inner.declarations.values() {
                    if existing.id != id && existing.declaration_no == *new_no {
                        return Err(AppError::DeclarationNoAlreadyExists(new_no.clone()));
                    }
                }
            }
        }
    }

    let decl = inner.declarations.get_mut(&id).unwrap();
    if let Some(new_no) = req.declaration_no {
        decl.declaration_no = new_no;
    }
    if let Some(items) = req.items {
        decl.items = items;
    }

    Ok(decl.clone())
}

pub fn delete_declaration(state: &AppState, id: Uuid) -> Result<(), AppError> {
    let mut inner = state.inner.lock().map_err(|e| AppError::Internal(e.to_string()))?;
    if !inner.declarations.contains_key(&id) {
        return Err(AppError::DeclarationNotFound(id.to_string()));
    }
    inner.declarations.remove(&id);
    Ok(())
}

pub fn merge_declarations(state: &AppState, req: MergeDeclarationRequest) -> Result<Declaration, AppError> {
    let mut inner = state.inner.lock().map_err(|e| AppError::Internal(e.to_string()))?;

    if req.declaration_ids.len() < 2 {
        return Err(AppError::MergeRequiresAtLeastTwo);
    }

    for existing in inner.declarations.values() {
        if existing.declaration_no == req.new_declaration_no {
            return Err(AppError::DeclarationNoAlreadyExists(req.new_declaration_no));
        }
    }

    let mut source_declarations = Vec::new();
    for id in &req.declaration_ids {
        let decl = inner
            .declarations
            .get(id)
            .ok_or_else(|| AppError::DeclarationNotFound(id.to_string()))?;
        if decl.status != DeclarationStatus::Draft {
            return Err(AppError::CannotMergeStatus(format!("{:?}", decl.status)));
        }
        source_declarations.push(decl.clone());
    }

    let mut merged_items: HashMap<String, DeclarationItem> = HashMap::new();
    for decl in &source_declarations {
        for item in &decl.items {
            let key = item.hs_code.clone();
            if let Some(existing) = merged_items.get_mut(&key) {
                if existing.name != item.name {
                    return Err(AppError::HsCodeNameMismatch(
                        key,
                        existing.name.clone(),
                        item.name.clone(),
                    ));
                }
                if existing.currency != item.currency {
                    return Err(AppError::InvalidInput(format!(
                        "HS code {} has different currencies: {:?} vs {:?}",
                        key, existing.currency, item.currency
                    )));
                }
                existing.quantity += item.quantity;
                let total_qty = existing.quantity;
                let total_amount = existing.amount() + item.amount();
                existing.unit_price = total_amount / total_qty;
            } else {
                merged_items.insert(key, item.clone());
            }
        }
    }

    let mut items: Vec<DeclarationItem> = merged_items.into_values().collect();
    items.sort_by(|a, b| a.hs_code.cmp(&b.hs_code));

    let new_decl = Declaration {
        id: Uuid::new_v4(),
        declaration_no: req.new_declaration_no,
        items,
        status: DeclarationStatus::Draft,
        declare_date: None,
        total_cny: None,
    };

    for id in &req.declaration_ids {
        if let Some(decl) = inner.declarations.get_mut(id) {
            decl.status = DeclarationStatus::Merged;
        }
    }

    inner.declarations.insert(new_decl.id, new_decl.clone());
    Ok(new_decl)
}

pub fn declare(state: &AppState, req: DeclareRequest) -> Result<Declaration, AppError> {
    let mut inner = state.inner.lock().map_err(|e| AppError::Internal(e.to_string()))?;

    {
        let decl = inner
            .declarations
            .get(&req.declaration_id)
            .ok_or_else(|| AppError::DeclarationNotFound(req.declaration_id.to_string()))?;

        if decl.status != DeclarationStatus::Draft {
            return Err(AppError::InvalidStatus {
                expected: "Draft".to_string(),
                actual: format!("{:?}", decl.status),
            });
        }

        let currencies = decl.currencies();
        for currency in currencies {
            if currency == Currency::CNY {
                continue;
            }
            let key = (req.declare_date, currency);
            if !inner.exchange_rates.contains_key(&key) {
                return Err(AppError::ExchangeRateNotFound(
                    format!("{:?}", currency),
                    req.declare_date.to_string(),
                ));
            }
        }
    }

    let mut total_cny = Decimal::ZERO;
    let items_to_process: Vec<DeclarationItem> = inner
        .declarations
        .get(&req.declaration_id)
        .unwrap()
        .items
        .clone();
    
    for item in &items_to_process {
        let amount = item.amount();
        if item.currency == Currency::CNY {
            total_cny += amount;
        } else {
            let key = (req.declare_date, item.currency);
            let rate = inner
                .exchange_rates
                .get(&key)
                .unwrap()
                .rate_to_cny;
            total_cny += amount * rate;
        }
    }

    let decl = inner.declarations.get_mut(&req.declaration_id).unwrap();
    decl.status = DeclarationStatus::Declared;
    decl.declare_date = Some(req.declare_date);
    decl.total_cny = Some(total_cny);

    Ok(decl.clone())
}
