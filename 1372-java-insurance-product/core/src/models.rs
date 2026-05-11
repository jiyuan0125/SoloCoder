use chrono::NaiveDate;
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use std::fmt;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum ProductType {
    Accident,
    CriticalIllness,
    Life,
}

impl fmt::Display for ProductType {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ProductType::Accident => write!(f, "意外险"),
            ProductType::CriticalIllness => write!(f, "重疾险"),
            ProductType::Life => write!(f, "寿险"),
        }
    }
}

impl ProductType {
    pub fn min_age_days(&self) -> i64 {
        match self {
            ProductType::Accident => 18 * 365,
            ProductType::CriticalIllness => 30,
            ProductType::Life => 18 * 365,
        }
    }

    pub fn max_age_days(&self) -> i64 {
        match self {
            ProductType::Accident => 65 * 365,
            ProductType::CriticalIllness => 55 * 365,
            ProductType::Life => 60 * 365,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Product {
    pub code: String,
    pub name: String,
    pub product_type: ProductType,
    pub base_premium: Decimal,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Beneficiary {
    pub name: String,
    pub id_card: String,
    pub ratio: Decimal,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Insured {
    pub name: String,
    pub id_card: String,
    pub birth_date: NaiveDate,
    pub occupation_risk_level: u8,
    pub has_safety_training: bool,
}

impl Insured {
    pub fn effective_risk_level(&self) -> u8 {
        if self.has_safety_training && self.occupation_risk_level > 1 {
            self.occupation_risk_level - 1
        } else {
            self.occupation_risk_level
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PolicyApplication {
    pub insured: Insured,
    pub product_code: String,
    pub application_date: NaiveDate,
    pub beneficiaries: Vec<Beneficiary>,
    pub insured_amount: Decimal,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Policy {
    pub policy_number: String,
    pub insured: Insured,
    pub product: Product,
    pub application_date: NaiveDate,
    pub beneficiaries: Vec<Beneficiary>,
    pub insured_amount: Decimal,
    pub premium: Decimal,
}

#[derive(Debug, Clone, thiserror::Error, Serialize, Deserialize)]
pub enum ValidationError {
    #[error("被保险人年龄不符合要求，产品({product})要求年龄在{min_age}至{max_age}周岁之间")]
    AgeOutOfRange {
        product: String,
        min_age: i64,
        max_age: i64,
    },
    #[error("产品不存在: {0}")]
    ProductNotFound(String),
    #[error("职业等级{level}不能投保{product}")]
    OccupationNotAllowed { level: u8, product: String },
    #[error("受益人数量不能超过3个，当前{count}个")]
    TooManyBeneficiaries { count: usize },
    #[error("受益人比例之和必须为100%，当前为{current}%")]
    BeneficiaryRatioSumInvalid { current: Decimal },
    #[error("已存在该被保险人在该产品下的投保记录")]
    AlreadyInsured,
    #[error("投保金额必须大于0")]
    InsuredAmountInvalid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ValidationResult {
    pub valid: bool,
    pub errors: Vec<ValidationError>,
    pub premium: Option<Decimal>,
}
