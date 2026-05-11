use crate::models::*;
use crate::store::SharedStore;
use crate::errors::SystemError;
use chrono::{Duration, Utc};
use std::collections::HashMap;
use uuid::Uuid;

const LOCK_DURATION_MINUTES: i64 = 15;

pub fn place_order(
    store: &SharedStore,
    quotation_id: Uuid,
    non_recommended_reason: Option<String>,
) -> Result<PurchaseOrder, SystemError> {
    let mut store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;

    let quotation = store_guard.quotation_by_id(quotation_id)
        .ok_or_else(|| SystemError::QuotationNotFound(quotation_id.to_string()))?;

    let purchase_request = store_guard.purchase_request_by_id(quotation.purchase_request_id)
        .ok_or_else(|| SystemError::PurchaseRequestNotFound(quotation.purchase_request_id.to_string()))?;

    if !quotation.is_recommended {
        if non_recommended_reason.is_none() || non_recommended_reason.as_ref().unwrap().trim().is_empty() {
            return Err(SystemError::NonRecommendedSupplierWithoutReason);
        }
    }

    let supplier = store_guard.supplier_by_id(quotation.supplier_id)
        .ok_or_else(|| SystemError::SupplierNotFound(quotation.supplier_id.to_string()))?;

    if supplier.status != SupplierStatus::Active {
        return Err(SystemError::SupplierNotActive(supplier.name));
    }

    let mut needs_merge: HashMap<Uuid, f64> = HashMap::new();
    let mut can_direct_order = true;

    for item in &quotation.items {
        let pr_item = purchase_request.items.iter()
            .find(|i| i.ingredient_id == item.ingredient_id)
            .ok_or_else(|| SystemError::IngredientNotFound(item.ingredient_id.to_string()))?;

        let supplier_ingredient = supplier.ingredients.iter()
            .find(|si| si.ingredient_id == item.ingredient_id)
            .ok_or_else(|| SystemError::SupplierDoesNotOfferIngredient {
                supplier: supplier.name.clone(),
                ingredient: item.ingredient_id.to_string(),
            })?;

        if pr_item.requested_quantity < supplier_ingredient.min_order_quantity {
            needs_merge.insert(item.ingredient_id, supplier_ingredient.min_order_quantity - pr_item.requested_quantity);
            can_direct_order = false;
        }
    }

    let order_items: Vec<OrderItem> = quotation.items.iter()
        .map(|qi| {
            let pr_item = purchase_request.items.iter()
                .find(|i| i.ingredient_id == qi.ingredient_id)
                .unwrap();
            OrderItem {
                ingredient_id: qi.ingredient_id,
                quantity: pr_item.requested_quantity,
                unit_price: qi.unit_price,
                subtotal: pr_item.requested_quantity * qi.unit_price,
            }
        })
        .collect();

    if !can_direct_order {
        return create_or_join_merge_pool(
            &mut store_guard,
            &purchase_request,
            &quotation,
            &supplier,
            &order_items,
            &needs_merge,
            non_recommended_reason,
        );
    }

    lock_stock_and_create_order(
        &mut store_guard,
        &purchase_request,
        &quotation,
        &supplier,
        &order_items,
        false,
        None,
        non_recommended_reason,
    )
}

fn create_or_join_merge_pool(
    store_guard: &mut crate::store::InMemoryStore,
    purchase_request: &PurchaseRequest,
    quotation: &Quotation,
    supplier: &Supplier,
    order_items: &[OrderItem],
    needs_merge: &HashMap<Uuid, f64>,
    non_recommended_reason: Option<String>,
) -> Result<PurchaseOrder, SystemError> {
    let now = Utc::now();
    let mut merged_store_orders: Vec<Uuid> = Vec::new();
    let mut total_quantities: HashMap<Uuid, f64> = HashMap::new();

    for (ingredient_id, shortfall) in needs_merge {
        let existing_pool = store_guard.merged_order_pools.iter_mut().find(|p| {
            p.supplier_id == supplier.id && 
            p.ingredient_id == *ingredient_id &&
            p.expires_at > now
        });

        if let Some(pool) = existing_pool {
            pool.total_quantity += shortfall;
            pool.store_orders.insert(purchase_request.store_id, *shortfall);
            for (store_id, _) in &pool.store_orders {
                if !merged_store_orders.contains(store_id) {
                    merged_store_orders.push(*store_id);
                }
            }
            *total_quantities.entry(*ingredient_id).or_insert(0.0) += pool.total_quantity;
        } else {
            let pool = MergedOrderPool {
                pool_id: Uuid::new_v4(),
                supplier_id: supplier.id,
                ingredient_id: *ingredient_id,
                total_quantity: *shortfall,
                store_orders: {
                    let mut map = HashMap::new();
                    map.insert(purchase_request.store_id, *shortfall);
                    map
                },
                created_at: now,
                expires_at: now + Duration::hours(2),
            };
            store_guard.merged_order_pools.push(pool);
            *total_quantities.entry(*ingredient_id).or_insert(0.0) += shortfall;
        }
    }

    let all_merged = merged_store_orders.len() > 1;
    let merged_with = if merged_store_orders.len() > 1 {
        Some(merged_store_orders.clone())
    } else {
        None
    };

    let order = lock_stock_and_create_order(
        store_guard,
        purchase_request,
        quotation,
        supplier,
        order_items,
        all_merged,
        merged_with,
        non_recommended_reason,
    )?;

    Ok(order)
}

fn lock_stock_and_create_order(
    store_guard: &mut crate::store::InMemoryStore,
    purchase_request: &PurchaseRequest,
    quotation: &Quotation,
    supplier: &Supplier,
    order_items: &[OrderItem],
    is_merged: bool,
    merged_with: Option<Vec<Uuid>>,
    non_recommended_reason: Option<String>,
) -> Result<PurchaseOrder, SystemError> {
    let now = Utc::now();
    let lock_expires = now + Duration::minutes(LOCK_DURATION_MINUTES);
    let order_id = Uuid::new_v4();

    for item in order_items {
        let supplier_idx = store_guard.suppliers.iter()
            .position(|s| s.id == supplier.id)
            .ok_or_else(|| SystemError::SupplierNotFound(supplier.id.to_string()))?;

        let supplier_ingredient = store_guard.suppliers[supplier_idx].ingredients.iter_mut()
            .find(|si| si.ingredient_id == item.ingredient_id)
            .ok_or_else(|| SystemError::SupplierDoesNotOfferIngredient {
                supplier: supplier.name.clone(),
                ingredient: item.ingredient_id.to_string(),
            })?;

        if supplier_ingredient.stock_quantity < item.quantity {
            return Err(SystemError::InsufficientStock {
                available: supplier_ingredient.stock_quantity,
                requested: item.quantity,
            });
        }

        supplier_ingredient.stock_quantity -= item.quantity;

        if let Some(existing_lock) = store_guard.in_progress_locks.iter_mut().find(|l| {
            l.supplier_id == supplier.id && l.ingredient_id == item.ingredient_id
        }) {
            existing_lock.locked_quantity += item.quantity;
            existing_lock.lock_expires_at = lock_expires;
            existing_lock.order_ids.push(order_id);
        } else {
            store_guard.in_progress_locks.push(InProgressLock {
                supplier_id: supplier.id,
                ingredient_id: item.ingredient_id,
                locked_quantity: item.quantity,
                lock_expires_at: lock_expires,
                order_ids: vec![order_id],
            });
        }
    }

    let total_price: f64 = order_items.iter().map(|i| i.subtotal).sum();
    let shipping_fee = calculate_proportional_shipping(order_items, quotation, store_guard);

    let order = PurchaseOrder {
        id: order_id,
        purchase_request_id: purchase_request.id,
        quotation_id: quotation.id,
        store_id: purchase_request.store_id,
        supplier_id: supplier.id,
        items: order_items.to_vec(),
        total_price,
        shipping_fee,
        is_merged,
        merged_with,
        selected_non_recommended: !quotation.is_recommended,
        non_recommended_reason,
        status: OrderStatus::Pending,
        created_at: now,
    };

    if let Some(pr) = store_guard.purchase_requests.iter_mut().find(|p| p.id == purchase_request.id) {
        pr.status = PurchaseStatus::OrderPlaced;
        pr.updated_at = now;
    }

    let current_month = now.format("%Y-%m").to_string();
    let order_total = order.total_price + order.shipping_fee;
    
    store_guard.update_store_monthly_stats(purchase_request.store_id, &current_month, |stats| {
        stats.total_purchase_amount += order_total;
    });

    store_guard.purchase_orders.push(order.clone());

    Ok(order)
}

fn calculate_proportional_shipping(
    order_items: &[OrderItem],
    quotation: &Quotation,
    store_guard: &crate::store::InMemoryStore,
) -> f64 {
    if quotation.total_weight == 0.0 {
        return 0.0;
    }

    let order_weight: f64 = order_items.iter()
        .map(|item| {
            store_guard.ingredient_by_id(item.ingredient_id)
                .map(|i| i.weight_per_unit * item.quantity)
                .unwrap_or(0.0)
        })
        .sum();

    (order_weight / quotation.total_weight) * quotation.shipping_fee
}

pub fn confirm_order(store: &SharedStore, order_id: Uuid) -> Result<PurchaseOrder, SystemError> {
    let mut store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;

    let order_idx = store_guard.purchase_orders.iter()
        .position(|o| o.id == order_id)
        .ok_or_else(|| SystemError::OrderNotFound(order_id.to_string()))?;

    if store_guard.purchase_orders[order_idx].status != OrderStatus::Pending {
        return Err(SystemError::InvalidPurchaseStatus);
    }

    store_guard.purchase_orders[order_idx].status = OrderStatus::Confirmed;

    let order = store_guard.purchase_orders[order_idx].clone();
    
    for item in &order.items {
        store_guard.in_progress_locks.retain(|l| {
            !(l.supplier_id == order.supplier_id && 
              l.ingredient_id == item.ingredient_id &&
              l.order_ids.contains(&order.id))
        });
    }

    Ok(order)
}

pub fn get_order(store: &SharedStore, order_id: Uuid) -> Result<PurchaseOrder, SystemError> {
    let store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;
    store_guard.purchase_order_by_id(order_id)
        .ok_or_else(|| SystemError::OrderNotFound(order_id.to_string()))
}
