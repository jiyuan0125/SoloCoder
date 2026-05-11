use thiserror::Error;

#[derive(Error, Debug)]
pub enum BudgetError {
    #[error("Category not found: {0}")]
    CategoryNotFound(String),
    
    #[error("Parent category not found: {0}")]
    ParentNotFound(String),
    
    #[error("Non-leaf categories cannot have budget amount")]
    NonLeafCannotHaveBudget,
    
    #[error("Leaf category must have budget amount")]
    LeafMustHaveBudget,
    
    #[error("Cannot set budget on category with children")]
    CannotSetBudgetOnNonLeaf,
    
    #[error("Expense amount must be positive")]
    InvalidExpenseAmount,
    
    #[error("Invalid date range")]
    InvalidDateRange,
    
    #[error("Invalid threshold: warning ({warning}) must be less than critical ({critical})")]
    InvalidThreshold { warning: f64, critical: f64 },
}
