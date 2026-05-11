use crate::models::*;
use crate::store::SharedStore;
use crate::errors::SystemError;
use chrono::Utc;
use uuid::Uuid;

pub fn create_supplier(
    store: &SharedStore,
    name: String,
    contact: String,
    phone: String,
    ingredients: Vec<SupplierIngredient>,
) -> Result<Supplier, SystemError> {
    let mut store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;

    let supplier = Supplier {
        id: Uuid::new_v4(),
        name,
        contact,
        phone,
        status: SupplierStatus::Active,
        fulfillment_rate: 0.95,
        monthly_scores: Vec::new(),
        ingredients,
        created_at: Utc::now(),
        updated_at: Utc::now(),
    };

    store_guard.suppliers.push(supplier.clone());
    Ok(supplier)
}

pub fn list_suppliers(store: &SharedStore) -> Result<Vec<Supplier>, SystemError> {
    let store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;
    Ok(store_guard.suppliers.clone())
}

pub fn get_supplier(store: &SharedStore, supplier_id: Uuid) -> Result<Supplier, SystemError> {
    let store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;
    store_guard.supplier_by_id(supplier_id)
        .ok_or_else(|| SystemError::SupplierNotFound(supplier_id.to_string()))
}

pub fn update_supplier_score(
    store: &SharedStore,
    supplier_id: Uuid,
    month: String,
    score: f64,
) -> Result<Supplier, SystemError> {
    if score < 0.0 || score > 100.0 {
        return Err(SystemError::InvalidInput("Score must be between 0 and 100".to_string()));
    }

    let mut store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;

    let supplier_idx = store_guard.suppliers.iter()
        .position(|s| s.id == supplier_id)
        .ok_or_else(|| SystemError::SupplierNotFound(supplier_id.to_string()))?;

    let supplier = &mut store_guard.suppliers[supplier_idx];
    
    if let Some(existing) = supplier.monthly_scores.iter_mut().find(|ms| ms.month == month) {
        existing.score = score;
    } else {
        supplier.monthly_scores.push(MonthlyScore { month, score });
    }

    evaluate_supplier_status(supplier);
    supplier.updated_at = Utc::now();

    Ok(supplier.clone())
}

fn evaluate_supplier_status(supplier: &mut Supplier) {
    let recent_scores: Vec<f64> = supplier.monthly_scores
        .iter()
        .rev()
        .take(3)
        .map(|ms| ms.score)
        .collect();

    if recent_scores.len() >= 3 {
        let all_below_60 = recent_scores.iter().all(|&s| s < 60.0);
        if all_below_60 {
            supplier.status = SupplierStatus::Disqualified;
            return;
        }
    }

    if recent_scores.len() >= 2 {
        let last_two = &recent_scores[..2];
        let all_below_70 = last_two.iter().all(|&s| s < 70.0);
        if all_below_70 && supplier.status == SupplierStatus::Active {
            supplier.status = SupplierStatus::Observation;
        } else if !all_below_70 && supplier.status == SupplierStatus::Observation {
            supplier.status = SupplierStatus::Active;
        }
    }
}

pub fn create_ingredient(
    store: &SharedStore,
    name: String,
    unit: String,
    category: String,
    weight_per_unit: f64,
) -> Result<Ingredient, SystemError> {
    let mut store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;

    let ingredient = Ingredient {
        id: Uuid::new_v4(),
        name,
        unit,
        category,
        weight_per_unit,
        created_at: Utc::now(),
    };

    store_guard.ingredients.push(ingredient.clone());
    Ok(ingredient)
}

pub fn list_ingredients(store: &SharedStore) -> Result<Vec<Ingredient>, SystemError> {
    let store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;
    Ok(store_guard.ingredients.clone())
}

pub fn create_store(
    store: &SharedStore,
    name: String,
    address: String,
    contact: String,
    phone: String,
) -> Result<Store, SystemError> {
    let mut store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;

    let new_store = Store {
        id: Uuid::new_v4(),
        name,
        address,
        contact,
        phone,
        created_at: Utc::now(),
    };

    store_guard.stores.push(new_store.clone());
    Ok(new_store)
}

pub fn list_stores(store: &SharedStore) -> Result<Vec<Store>, SystemError> {
    let store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;
    Ok(store_guard.stores.clone())
}

pub fn create_purchase_request(
    store: &SharedStore,
    store_id: Uuid,
    purchase_type: PurchaseType,
    items: Vec<PurchaseItem>,
) -> Result<PurchaseRequest, SystemError> {
    let mut store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;

    if store_guard.store_by_id(store_id).is_none() {
        return Err(SystemError::StoreNotFound(store_id.to_string()));
    }

    for item in &items {
        if store_guard.ingredient_by_id(item.ingredient_id).is_none() {
            return Err(SystemError::IngredientNotFound(item.ingredient_id.to_string()));
        }
    }

    let pr = PurchaseRequest {
        id: Uuid::new_v4(),
        store_id,
        purchase_type,
        items,
        status: PurchaseStatus::Draft,
        created_at: Utc::now(),
        updated_at: Utc::now(),
    };

    store_guard.purchase_requests.push(pr.clone());
    Ok(pr)
}

pub fn list_purchase_requests(store: &SharedStore) -> Result<Vec<PurchaseRequest>, SystemError> {
    let store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;
    Ok(store_guard.purchase_requests.clone())
}

pub fn get_purchase_request(store: &SharedStore, id: Uuid) -> Result<PurchaseRequest, SystemError> {
    let store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;
    store_guard.purchase_request_by_id(id)
        .ok_or_else(|| SystemError::PurchaseRequestNotFound(id.to_string()))
}
