use crate::models::*;
use crate::store::SharedStore;
use crate::errors::SystemError;
use chrono::Utc;
use uuid::Uuid;

const PRICE_WEIGHT: f64 = 0.40;
const QUALITY_WEIGHT: f64 = 0.30;
const TIMELINESS_WEIGHT: f64 = 0.20;
const FULFILLMENT_WEIGHT: f64 = 0.10;

const SHIPPING_RATE_PER_KG: f64 = 2.0;

pub fn request_quotations(store: &SharedStore, purchase_request_id: Uuid) -> Result<Vec<Quotation>, SystemError> {
    let mut store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;

    let purchase = store_guard.purchase_request_by_id(purchase_request_id)
        .ok_or_else(|| SystemError::PurchaseRequestNotFound(purchase_request_id.to_string()))?;

    if purchase.status != PurchaseStatus::Draft && purchase.status != PurchaseStatus::InQuotation {
        return Err(SystemError::InvalidPurchaseStatus);
    }

    if purchase.purchase_type == PurchaseType::Emergency {
        return Err(SystemError::InvalidInput("Emergency purchases do not use quotation process".to_string()));
    }

    let active_suppliers: Vec<Supplier> = store_guard.suppliers
        .iter()
        .filter(|s| s.status == SupplierStatus::Active)
        .cloned()
        .collect();

    if active_suppliers.is_empty() {
        return Err(SystemError::SupplierNotFound("No active suppliers available".to_string()));
    }

    let mut quotations = Vec::new();

    for supplier in &active_suppliers {
        let mut can_quote_all = true;
        let mut quotation_items = Vec::new();
        let mut total_price = 0.0;
        let mut total_weight = 0.0;
        let mut max_delivery_time = 0;

        for item in &purchase.items {
            let supplier_ingredient = supplier.ingredients
                .iter()
                .find(|si| si.ingredient_id == item.ingredient_id);

            match supplier_ingredient {
                None => {
                    can_quote_all = false;
                    break;
                }
                Some(si) => {
                    let ingredient = store_guard.ingredient_by_id(item.ingredient_id)
                        .ok_or_else(|| SystemError::IngredientNotFound(item.ingredient_id.to_string()))?;

                    if si.stock_quantity < item.requested_quantity {
                        can_quote_all = false;
                        break;
                    }

                    let subtotal = si.unit_price * item.requested_quantity;
                    let weight = ingredient.weight_per_unit * item.requested_quantity;

                    quotation_items.push(QuotationItem {
                        ingredient_id: item.ingredient_id,
                        unit_price: si.unit_price,
                        available_quantity: si.stock_quantity,
                        quality_level: si.quality_level,
                        subtotal,
                    });

                    total_price += subtotal;
                    total_weight += weight;
                    if si.delivery_time_hours > max_delivery_time {
                        max_delivery_time = si.delivery_time_hours;
                    }
                }
            }
        }

        if can_quote_all {
            let shipping_fee = total_weight * SHIPPING_RATE_PER_KG;
            
            let quotation = Quotation {
                id: Uuid::new_v4(),
                purchase_request_id: purchase.id,
                supplier_id: supplier.id,
                items: quotation_items,
                total_price,
                total_weight,
                shipping_fee,
                delivery_time_hours: max_delivery_time,
                composite_score: 0.0,
                is_recommended: false,
                created_at: Utc::now(),
            };

            quotations.push(quotation);
        }
    }

    if quotations.is_empty() {
        return Err(SystemError::QuotationNotAvailable);
    }

    let price_scores = calculate_normalized_scores(&quotations, |q| q.total_price, true);
    let quality_scores = calculate_quality_scores(&quotations);
    let timeliness_scores = calculate_normalized_scores(&quotations, |q| q.delivery_time_hours as f64, true);
    let fulfillment_scores = calculate_fulfillment_scores(&quotations, &store_guard.suppliers);

    let mut scored_quotations = Vec::new();
    for (idx, q) in quotations.iter().enumerate() {
        let total_score =
            price_scores[idx] * PRICE_WEIGHT +
            quality_scores[idx] * QUALITY_WEIGHT +
            timeliness_scores[idx] * TIMELINESS_WEIGHT +
            fulfillment_scores[idx] * FULFILLMENT_WEIGHT;

        let mut scored = q.clone();
        scored.composite_score = total_score;
        scored_quotations.push(scored);
    }

    scored_quotations.sort_by(|a, b| b.composite_score.partial_cmp(&a.composite_score).unwrap());

    if let Some(best) = scored_quotations.first_mut() {
        best.is_recommended = true;
    }

    for q in &scored_quotations {
        store_guard.quotations.push(q.clone());
    }

    if let Some(pr) = store_guard.purchase_requests.iter_mut().find(|p| p.id == purchase_request_id) {
        pr.status = PurchaseStatus::QuotationCompleted;
        pr.updated_at = Utc::now();
    }

    Ok(scored_quotations)
}

fn calculate_normalized_scores<F: Fn(&Quotation) -> f64>(
    quotations: &[Quotation],
    get_value: F,
    lower_is_better: bool,
) -> Vec<f64> {
    if quotations.is_empty() {
        return vec![];
    }

    let values: Vec<f64> = quotations.iter().map(&get_value).collect();
    let min_val = *values.iter().fold(&f64::INFINITY, |a, b| if a < b { a } else { b });
    let max_val = *values.iter().fold(&f64::NEG_INFINITY, |a, b| if a > b { a } else { b });

    if max_val == min_val {
        return vec![100.0; quotations.len()];
    }

    values.iter().map(|&v| {
        let normalized = (v - min_val) / (max_val - min_val);
        if lower_is_better {
            100.0 * (1.0 - normalized)
        } else {
            100.0 * normalized
        }
    }).collect()
}

fn calculate_quality_scores(quotations: &[Quotation]) -> Vec<f64> {
    quotations.iter().map(|q| {
        let avg_quality: f64 = q.items.iter()
            .map(|item| item.quality_level as i32 as f64)
            .sum::<f64>() / q.items.len() as f64;
        avg_quality * 20.0
    }).collect()
}

fn calculate_fulfillment_scores(quotations: &[Quotation], suppliers: &[Supplier]) -> Vec<f64> {
    quotations.iter().map(|q| {
        suppliers.iter()
            .find(|s| s.id == q.supplier_id)
            .map(|s| s.fulfillment_rate * 100.0)
            .unwrap_or(50.0)
    }).collect()
}

pub fn list_quotations(store: &SharedStore, purchase_request_id: Uuid) -> Result<Vec<Quotation>, SystemError> {
    let store_guard = store.lock().map_err(|e| SystemError::Internal(format!("Lock error: {}", e)))?;
    Ok(store_guard.quotations_by_request(purchase_request_id))
}
