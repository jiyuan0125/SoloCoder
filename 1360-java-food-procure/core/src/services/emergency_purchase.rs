use crate::models::*;
use crate::store::SharedStore;
use crate::errors::SystemError;
use chrono::Utc;
use uuid::Uuid;

const EMERGENCY_SINGLE_LIMIT: f64 = 2000.0;
const EMERGENCY_MONTHLY_PERCENTAGE: f64 = 0.10;

pub fn place_emergency_order(
    store: &SharedStore,
    store_id: Uuid,
    supplier_id: Uuid,
    items: Vec<PurchaseItem>,
) -> Result<PurchaseOrder, SystemError> {
    let mut store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;

    if store_guard.store_by_id(store_id).is_none() {
        return Err(SystemError::StoreNotFound(store_id.to_string()));
    }

    let supplier = store_guard.supplier_by_id(supplier_id)
        .ok_or_else(|| SystemError::SupplierNotFound(supplier_id.to_string()))?;

    if supplier.status != SupplierStatus::Active {
        return Err(SystemError::SupplierNotActive(supplier.name));
    }

    let mut order_items = Vec::new();
    let mut total_price = 0.0;

    for item in &items {
        let supplier_ingredient = supplier.ingredients.iter()
            .find(|si| si.ingredient_id == item.ingredient_id)
            .ok_or_else(|| SystemError::SupplierDoesNotOfferIngredient {
                supplier: supplier.name.clone(),
                ingredient: item.ingredient_id.to_string(),
            })?;

        if supplier_ingredient.stock_quantity < item.requested_quantity {
            return Err(SystemError::InsufficientStock {
                available: supplier_ingredient.stock_quantity,
                requested: item.requested_quantity,
            });
        }

        let subtotal = supplier_ingredient.unit_price * item.requested_quantity;
        total_price += subtotal;

        order_items.push(OrderItem {
            ingredient_id: item.ingredient_id,
            quantity: item.requested_quantity,
            unit_price: supplier_ingredient.unit_price,
            subtotal,
        });
    }

    if total_price > EMERGENCY_SINGLE_LIMIT {
        return Err(SystemError::EmergencyPurchaseSingleLimitExceeded(total_price));
    }

    let current_month = Utc::now().format("%Y-%m").to_string();
    let stats = store_guard.get_or_create_store_monthly_stats(store_id, &current_month);
    
    let new_emergency_total = stats.emergency_purchase_amount + total_price;
    let new_total_amount = stats.total_purchase_amount + total_price;
    
    let emergency_percentage = if new_total_amount > 0.0 {
        new_emergency_total / new_total_amount
    } else {
        0.0
    };

    if emergency_percentage > EMERGENCY_MONTHLY_PERCENTAGE {
        return Err(SystemError::EmergencyPurchaseMonthlyLimitExceeded(emergency_percentage * 100.0));
    }

    let pr = PurchaseRequest {
        id: Uuid::new_v4(),
        store_id,
        purchase_type: PurchaseType::Emergency,
        items,
        status: PurchaseStatus::OrderPlaced,
        created_at: Utc::now(),
        updated_at: Utc::now(),
    };

    let order = PurchaseOrder {
        id: Uuid::new_v4(),
        purchase_request_id: pr.id,
        quotation_id: Uuid::nil(),
        store_id,
        supplier_id,
        items: order_items.clone(),
        total_price,
        shipping_fee: 0.0,
        is_merged: false,
        merged_with: None,
        selected_non_recommended: false,
        non_recommended_reason: None,
        status: OrderStatus::Pending,
        created_at: Utc::now(),
    };

    let supplier_idx = store_guard.suppliers.iter()
        .position(|s| s.id == supplier_id)
        .unwrap();

    for item in &order_items {
        let si = store_guard.suppliers[supplier_idx].ingredients.iter_mut()
            .find(|si| si.ingredient_id == item.ingredient_id)
            .unwrap();
        si.stock_quantity -= item.quantity;
    }

    store_guard.update_store_monthly_stats(store_id, &current_month, |stats| {
        stats.total_purchase_amount += total_price;
        stats.emergency_purchase_amount += total_price;
        stats.emergency_purchase_count += 1;
    });

    store_guard.purchase_requests.push(pr);
    store_guard.purchase_orders.push(order.clone());

    Ok(order)
}

pub fn get_store_emergency_stats(
    store: &SharedStore,
    store_id: Uuid,
    month: Option<String>,
) -> Result<StoreMonthlyStats, SystemError> {
    let mut store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;
    
    let month_key = month.unwrap_or_else(|| Utc::now().format("%Y-%m").to_string());
    
    Ok(store_guard.get_or_create_store_monthly_stats(store_id, &month_key))
}
