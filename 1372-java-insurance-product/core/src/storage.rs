use crate::models::*;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

#[derive(Clone)]
pub struct InMemoryStore {
    products: Arc<Mutex<HashMap<String, Product>>>,
    policies: Arc<Mutex<HashMap<String, Policy>>>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        let store = Self {
            products: Arc::new(Mutex::new(HashMap::new())),
            policies: Arc::new(Mutex::new(HashMap::new())),
        };
        store.initialize_products();
        store
    }

    fn initialize_products(&self) {
        let products = [
            Product {
                code: "ACC001".to_string(),
                name: "综合意外险".to_string(),
                product_type: ProductType::Accident,
                base_premium: rust_decimal_macros::dec!(50),
            },
            Product {
                code: "CI001".to_string(),
                name: "重大疾病险".to_string(),
                product_type: ProductType::CriticalIllness,
                base_premium: rust_decimal_macros::dec!(200),
            },
            Product {
                code: "LIFE001".to_string(),
                name: "终身寿险".to_string(),
                product_type: ProductType::Life,
                base_premium: rust_decimal_macros::dec!(150),
            },
        ];

        let mut product_map = self.products.lock().unwrap();
        for product in products {
            product_map.insert(product.code.clone(), product);
        }
    }

    pub fn get_product(&self, code: &str) -> Option<Product> {
        self.products.lock().unwrap().get(code).cloned()
    }

    pub fn list_products(&self) -> Vec<Product> {
        self.products
            .lock()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    pub fn has_existing_policy(&self, id_card: &str, product_code: &str) -> bool {
        self.policies
            .lock()
            .unwrap()
            .values()
            .any(|p| p.insured.id_card == id_card && p.product.code == product_code)
    }

    pub fn add_policy(&self, policy: Policy) {
        self.policies
            .lock()
            .unwrap()
            .insert(policy.policy_number.clone(), policy);
    }

    pub fn list_policies(&self) -> Vec<Policy> {
        self.policies
            .lock()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    pub fn get_policy(&self, policy_number: &str) -> Option<Policy> {
        self.policies.lock().unwrap().get(policy_number).cloned()
    }
}

impl Default for InMemoryStore {
    fn default() -> Self {
        Self::new()
    }
}
