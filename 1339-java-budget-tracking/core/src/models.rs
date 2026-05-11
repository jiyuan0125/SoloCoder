use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};
use uuid::Uuid;
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BudgetCategory {
    pub id: Uuid,
    pub name: String,
    pub parent_id: Option<Uuid>,
    pub budget_amount: Option<f64>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BudgetCategoryTree {
    pub id: Uuid,
    pub name: String,
    pub parent_id: Option<Uuid>,
    pub budget_amount: Option<f64>,
    pub computed_budget: f64,
    pub children: Vec<BudgetCategoryTree>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Expense {
    pub id: Uuid,
    pub category_id: Uuid,
    pub amount: f64,
    pub date: DateTime<Utc>,
    pub description: String,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateBudgetCategoryRequest {
    pub name: String,
    pub parent_id: Option<Uuid>,
    pub budget_amount: Option<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UpdateBudgetCategoryRequest {
    pub name: Option<String>,
    pub budget_amount: Option<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateExpenseRequest {
    pub category_id: Uuid,
    pub amount: f64,
    pub date: DateTime<Utc>,
    pub description: String,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum AlertLevel {
    None,
    Warning,
    Critical,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionSummary {
    pub category_id: Uuid,
    pub category_name: String,
    pub budget_amount: f64,
    pub actual_spent: f64,
    pub execution_rate: f64,
    pub alert_level: AlertLevel,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AlertConfig {
    pub warning_threshold: f64,
    pub critical_threshold: f64,
}

impl Default for AlertConfig {
    fn default() -> Self {
        AlertConfig {
            warning_threshold: 90.0,
            critical_threshold: 100.0,
        }
    }
}

impl BudgetCategoryTree {
    pub fn from_categories(categories: &[BudgetCategory]) -> Vec<Self> {
        let mut category_map: HashMap<Uuid, BudgetCategoryTree> = HashMap::new();
        for cat in categories {
            category_map.insert(
                cat.id,
                BudgetCategoryTree {
                    id: cat.id,
                    name: cat.name.clone(),
                    parent_id: cat.parent_id,
                    budget_amount: cat.budget_amount,
                    computed_budget: cat.budget_amount.unwrap_or(0.0),
                    children: Vec::new(),
                },
            );
        }

        let mut children_map: HashMap<Uuid, Vec<Uuid>> = HashMap::new();
        let mut roots: Vec<Uuid> = Vec::new();
        
        for cat in categories {
            if let Some(parent_id) = cat.parent_id {
                children_map.entry(parent_id).or_default().push(cat.id);
            } else {
                roots.push(cat.id);
            }
        }

        for (parent_id, child_ids) in children_map {
            let mut children_to_add = Vec::new();
            for child_id in child_ids {
                if let Some(child) = category_map.remove(&child_id) {
                    children_to_add.push(child);
                }
            }
            if let Some(parent) = category_map.get_mut(&parent_id) {
                parent.children.extend(children_to_add);
            }
        }

        let mut result: Vec<BudgetCategoryTree> = Vec::new();
        for root_id in roots {
            if let Some(root) = category_map.remove(&root_id) {
                result.push(root);
            }
        }

        Self::compute_budgets(&mut result);
        result
    }

    fn compute_budgets(nodes: &mut [BudgetCategoryTree]) {
        for node in nodes {
            Self::compute_budgets(&mut node.children);
            if !node.children.is_empty() {
                node.computed_budget = node.children.iter().map(|c| c.computed_budget).sum();
                node.budget_amount = None;
            }
        }
    }

    pub fn find_category(&self, id: Uuid) -> Option<&Self> {
        if self.id == id {
            return Some(self);
        }
        for child in &self.children {
            if let Some(found) = child.find_category(id) {
                return Some(found);
            }
        }
        None
    }

    pub fn find_category_mut(&mut self, id: Uuid) -> Option<&mut Self> {
        if self.id == id {
            return Some(self);
        }
        for child in &mut self.children {
            if let Some(found) = child.find_category_mut(id) {
                return Some(found);
            }
        }
        None
    }
}
