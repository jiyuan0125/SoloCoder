use std::sync::RwLock;

use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::error::BudgetError;
use crate::models::{
    AlertConfig, AlertLevel, BudgetCategory, BudgetCategoryTree, CreateBudgetCategoryRequest,
    CreateExpenseRequest, ExecutionSummary, Expense, UpdateBudgetCategoryRequest,
};

pub struct BudgetService {
    categories: RwLock<Vec<BudgetCategory>>,
    expenses: RwLock<Vec<Expense>>,
    alert_config: RwLock<AlertConfig>,
}

impl Default for BudgetService {
    fn default() -> Self {
        Self::new()
    }
}

impl BudgetService {
    pub fn new() -> Self {
        BudgetService {
            categories: RwLock::new(Vec::new()),
            expenses: RwLock::new(Vec::new()),
            alert_config: RwLock::new(AlertConfig::default()),
        }
    }

    pub fn create_category(
        &self,
        req: CreateBudgetCategoryRequest,
    ) -> Result<BudgetCategory, BudgetError> {
        let categories = self.categories.read().unwrap();

        if let Some(parent_id) = req.parent_id {
            if !categories.iter().any(|c| c.id == parent_id) {
                return Err(BudgetError::ParentNotFound(parent_id.to_string()));
            }
        }

        if req.parent_id.is_some() && req.budget_amount.is_none() {
            return Err(BudgetError::LeafMustHaveBudget);
        }

        drop(categories);

        let category = BudgetCategory {
            id: Uuid::new_v4(),
            name: req.name,
            parent_id: req.parent_id,
            budget_amount: req.budget_amount,
            created_at: Utc::now(),
        };

        let mut categories = self.categories.write().unwrap();
        if let Some(parent_id) = req.parent_id {
            if let Some(parent) = categories.iter_mut().find(|c| c.id == parent_id) {
                if parent.budget_amount.is_some() {
                    parent.budget_amount = None;
                }
            }
        }
        categories.push(category.clone());

        Ok(category)
    }

    pub fn list_categories(&self) -> Vec<BudgetCategory> {
        self.categories.read().unwrap().clone()
    }

    pub fn get_category_tree(&self) -> Vec<BudgetCategoryTree> {
        let categories = self.categories.read().unwrap();
        BudgetCategoryTree::from_categories(&categories)
    }

    pub fn update_category(
        &self,
        id: Uuid,
        req: UpdateBudgetCategoryRequest,
    ) -> Result<BudgetCategory, BudgetError> {
        let mut categories = self.categories.write().unwrap();

        let idx = categories
            .iter()
            .position(|c| c.id == id)
            .ok_or_else(|| BudgetError::CategoryNotFound(id.to_string()))?;

        let has_children = categories.iter().any(|c| c.parent_id == Some(id));

        if has_children && req.budget_amount.is_some() {
            return Err(BudgetError::CannotSetBudgetOnNonLeaf);
        }

        if let Some(name) = req.name {
            categories[idx].name = name;
        }

        if req.budget_amount.is_some() {
            categories[idx].budget_amount = req.budget_amount;
        }

        Ok(categories[idx].clone())
    }

    pub fn delete_category(&self, id: Uuid) -> Result<(), BudgetError> {
        let mut categories = self.categories.write().unwrap();

        if categories.iter().any(|c| c.parent_id == Some(id)) {
            return Err(BudgetError::CannotSetBudgetOnNonLeaf);
        }

        let original_len = categories.len();
        categories.retain(|c| c.id != id);

        if categories.len() == original_len {
            return Err(BudgetError::CategoryNotFound(id.to_string()));
        }

        Ok(())
    }

    pub fn create_expense(&self, req: CreateExpenseRequest) -> Result<Expense, BudgetError> {
        if req.amount <= 0.0 {
            return Err(BudgetError::InvalidExpenseAmount);
        }

        let categories = self.categories.read().unwrap();
        if !categories.iter().any(|c| c.id == req.category_id) {
            return Err(BudgetError::CategoryNotFound(req.category_id.to_string()));
        }
        drop(categories);

        let expense = Expense {
            id: Uuid::new_v4(),
            category_id: req.category_id,
            amount: req.amount,
            date: req.date,
            description: req.description,
            created_at: Utc::now(),
        };

        let mut expenses = self.expenses.write().unwrap();
        expenses.push(expense.clone());

        Ok(expense)
    }

    pub fn list_expenses(&self) -> Vec<Expense> {
        self.expenses.read().unwrap().clone()
    }

    pub fn get_alert_config(&self) -> AlertConfig {
        self.alert_config.read().unwrap().clone()
    }

    pub fn set_alert_config(&self, config: AlertConfig) -> Result<(), BudgetError> {
        if config.warning_threshold >= config.critical_threshold {
            return Err(BudgetError::InvalidThreshold {
                warning: config.warning_threshold,
                critical: config.critical_threshold,
            });
        }
        *self.alert_config.write().unwrap() = config;
        Ok(())
    }

    fn collect_leaf_category_ids(node: &BudgetCategoryTree, leaves: &mut Vec<Uuid>) {
        if node.children.is_empty() {
            leaves.push(node.id);
        } else {
            for child in &node.children {
                Self::collect_leaf_category_ids(child, leaves);
            }
        }
    }

    pub fn get_execution_summary(
        &self,
        category_id: Option<Uuid>,
        start_date: Option<DateTime<Utc>>,
        end_date: Option<DateTime<Utc>>,
    ) -> Result<Vec<ExecutionSummary>, BudgetError> {
        if let (Some(start), Some(end)) = (start_date, end_date) {
            if start > end {
                return Err(BudgetError::InvalidDateRange);
            }
        }

        let tree = self.get_category_tree();
        let alert_config = self.get_alert_config();
        let expenses = self.expenses.read().unwrap().clone();

        let mut summaries = Vec::new();

        if let Some(cat_id) = category_id {
            let mut found = false;
            for root in &tree {
                if let Some(node) = root.find_category(cat_id) {
                    found = true;
                    summaries.push(self.build_summary(
                        node,
                        &expenses,
                        start_date,
                        end_date,
                        &alert_config,
                    ));
                    break;
                }
            }
            if !found {
                return Err(BudgetError::CategoryNotFound(cat_id.to_string()));
            }
        } else {
            for root in &tree {
                summaries.push(self.build_summary(
                    root,
                    &expenses,
                    start_date,
                    end_date,
                    &alert_config,
                ));
            }
        }

        Ok(summaries)
    }

    fn build_summary(
        &self,
        node: &BudgetCategoryTree,
        expenses: &[Expense],
        start_date: Option<DateTime<Utc>>,
        end_date: Option<DateTime<Utc>>,
        alert_config: &AlertConfig,
    ) -> ExecutionSummary {
        let mut leaf_ids = Vec::new();
        Self::collect_leaf_category_ids(node, &mut leaf_ids);

        let actual_spent: f64 = expenses
            .iter()
            .filter(|e| leaf_ids.contains(&e.category_id))
            .filter(|e| {
                if let Some(start) = start_date {
                    if e.date < start {
                        return false;
                    }
                }
                if let Some(end) = end_date {
                    if e.date > end {
                        return false;
                    }
                }
                true
            })
            .map(|e| e.amount)
            .sum();

        let budget_amount = node.computed_budget;
        let execution_rate = if budget_amount > 0.0 {
            (actual_spent / budget_amount) * 100.0
        } else {
            0.0
        };

        let alert_level = if execution_rate >= alert_config.critical_threshold {
            AlertLevel::Critical
        } else if execution_rate >= alert_config.warning_threshold {
            AlertLevel::Warning
        } else {
            AlertLevel::None
        };

        ExecutionSummary {
            category_id: node.id,
            category_name: node.name.clone(),
            budget_amount,
            actual_spent,
            execution_rate,
            alert_level,
        }
    }

    pub fn get_full_tree_with_execution(
        &self,
        start_date: Option<DateTime<Utc>>,
        end_date: Option<DateTime<Utc>>,
    ) -> Vec<(BudgetCategoryTree, ExecutionSummary)> {
        let tree = self.get_category_tree();
        let alert_config = self.get_alert_config();
        let expenses = self.expenses.read().unwrap().clone();

        let mut result = Vec::new();
        for root in &tree {
            self.build_tree_with_execution_recursive(
                root,
                &expenses,
                start_date,
                end_date,
                &alert_config,
                &mut result,
            );
        }
        result
    }

    fn build_tree_with_execution_recursive(
        &self,
        node: &BudgetCategoryTree,
        expenses: &[Expense],
        start_date: Option<DateTime<Utc>>,
        end_date: Option<DateTime<Utc>>,
        alert_config: &AlertConfig,
        result: &mut Vec<(BudgetCategoryTree, ExecutionSummary)>,
    ) {
        let summary = self.build_summary(node, expenses, start_date, end_date, alert_config);
        result.push((node.clone(), summary));

        for child in &node.children {
            self.build_tree_with_execution_recursive(
                child,
                expenses,
                start_date,
                end_date,
                alert_config,
                result,
            );
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::Duration;

    #[test]
    fn test_create_leaf_category_with_budget() {
        let service = BudgetService::new();
        let req = CreateBudgetCategoryRequest {
            name: "办公用品".to_string(),
            parent_id: None,
            budget_amount: Some(10000.0),
        };
        let category = service.create_category(req).unwrap();
        assert_eq!(category.name, "办公用品");
        assert_eq!(category.budget_amount, Some(10000.0));
    }

    #[test]
    fn test_create_tree_structure() {
        let service = BudgetService::new();
        
        let parent = service
            .create_category(CreateBudgetCategoryRequest {
                name: "办公费用".to_string(),
                parent_id: None,
                budget_amount: None,
            })
            .unwrap();

        let child1 = service
            .create_category(CreateBudgetCategoryRequest {
                name: "办公用品".to_string(),
                parent_id: Some(parent.id),
                budget_amount: Some(5000.0),
            })
            .unwrap();

        let child2 = service
            .create_category(CreateBudgetCategoryRequest {
                name: "水电费".to_string(),
                parent_id: Some(parent.id),
                budget_amount: Some(3000.0),
            })
            .unwrap();

        let tree = service.get_category_tree();
        assert_eq!(tree.len(), 1);
        
        let root = &tree[0];
        assert_eq!(root.children.len(), 2);
        assert_eq!(root.computed_budget, 8000.0);
    }

    #[test]
    fn test_execution_rate_calculation() {
        let service = BudgetService::new();
        
        let category = service
            .create_category(CreateBudgetCategoryRequest {
                name: "办公用品".to_string(),
                parent_id: None,
                budget_amount: Some(10000.0),
            })
            .unwrap();

        service
            .create_expense(CreateExpenseRequest {
                category_id: category.id,
                amount: 5000.0,
                date: Utc::now(),
                description: "购买文具".to_string(),
            })
            .unwrap();

        let summaries = service
            .get_execution_summary(Some(category.id), None, None)
            .unwrap();
        
        assert_eq!(summaries.len(), 1);
        assert_eq!(summaries[0].actual_spent, 5000.0);
        assert_eq!(summaries[0].execution_rate, 50.0);
        assert_eq!(summaries[0].alert_level, AlertLevel::None);
    }

    #[test]
    fn test_execution_rate_with_date_filter() {
        let service = BudgetService::new();
        
        let category = service
            .create_category(CreateBudgetCategoryRequest {
                name: "办公用品".to_string(),
                parent_id: None,
                budget_amount: Some(10000.0),
            })
            .unwrap();

        let now = Utc::now();
        let last_month = now - Duration::days(30);
        let next_month = now + Duration::days(30);

        service
            .create_expense(CreateExpenseRequest {
                category_id: category.id,
                amount: 3000.0,
                date: last_month,
                description: "上月支出".to_string(),
            })
            .unwrap();

        service
            .create_expense(CreateExpenseRequest {
                category_id: category.id,
                amount: 4000.0,
                date: now,
                description: "本月支出".to_string(),
            })
            .unwrap();

        let summaries = service
            .get_execution_summary(Some(category.id), Some(now - Duration::days(1)), Some(next_month))
            .unwrap();
        
        assert_eq!(summaries[0].actual_spent, 4000.0);
        assert_eq!(summaries[0].budget_amount, 10000.0);
        assert_eq!(summaries[0].execution_rate, 40.0);
    }

    #[test]
    fn test_alert_levels() {
        let service = BudgetService::new();
        
        let category = service
            .create_category(CreateBudgetCategoryRequest {
                name: "办公用品".to_string(),
                parent_id: None,
                budget_amount: Some(10000.0),
            })
            .unwrap();

        service
            .create_expense(CreateExpenseRequest {
                category_id: category.id,
                amount: 9500.0,
                date: Utc::now(),
                description: "大额支出".to_string(),
            })
            .unwrap();

        let summaries = service
            .get_execution_summary(Some(category.id), None, None)
            .unwrap();
        
        assert_eq!(summaries[0].alert_level, AlertLevel::Warning);

        service
            .create_expense(CreateExpenseRequest {
                category_id: category.id,
                amount: 1000.0,
                date: Utc::now(),
                description: "超预算支出".to_string(),
            })
            .unwrap();

        let summaries = service
            .get_execution_summary(Some(category.id), None, None)
            .unwrap();
        
        assert_eq!(summaries[0].alert_level, AlertLevel::Critical);
    }

    #[test]
    fn test_custom_alert_config() {
        let service = BudgetService::new();
        
        service
            .set_alert_config(AlertConfig {
                warning_threshold: 70.0,
                critical_threshold: 90.0,
            })
            .unwrap();

        let category = service
            .create_category(CreateBudgetCategoryRequest {
                name: "办公用品".to_string(),
                parent_id: None,
                budget_amount: Some(10000.0),
            })
            .unwrap();

        service
            .create_expense(CreateExpenseRequest {
                category_id: category.id,
                amount: 7500.0,
                date: Utc::now(),
                description: "测试".to_string(),
            })
            .unwrap();

        let summaries = service
            .get_execution_summary(Some(category.id), None, None)
            .unwrap();
        
        assert_eq!(summaries[0].alert_level, AlertLevel::Warning);
    }

    #[test]
    fn test_hierarchy_execution_summary() {
        let service = BudgetService::new();
        
        let parent = service
            .create_category(CreateBudgetCategoryRequest {
                name: "办公费用".to_string(),
                parent_id: None,
                budget_amount: None,
            })
            .unwrap();

        let child1 = service
            .create_category(CreateBudgetCategoryRequest {
                name: "办公用品".to_string(),
                parent_id: Some(parent.id),
                budget_amount: Some(5000.0),
            })
            .unwrap();

        let child2 = service
            .create_category(CreateBudgetCategoryRequest {
                name: "水电费".to_string(),
                parent_id: Some(parent.id),
                budget_amount: Some(3000.0),
            })
            .unwrap();

        service
            .create_expense(CreateExpenseRequest {
                category_id: child1.id,
                amount: 3000.0,
                date: Utc::now(),
                description: "文具".to_string(),
            })
            .unwrap();

        service
            .create_expense(CreateExpenseRequest {
                category_id: child2.id,
                amount: 2000.0,
                date: Utc::now(),
                description: "电费".to_string(),
            })
            .unwrap();

        let parent_summary = service
            .get_execution_summary(Some(parent.id), None, None)
            .unwrap();
        
        assert_eq!(parent_summary[0].budget_amount, 8000.0);
        assert_eq!(parent_summary[0].actual_spent, 5000.0);
        assert_eq!(parent_summary[0].execution_rate, 62.5);
    }
}
