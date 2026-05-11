use crate::models::*;
use rust_decimal::{prelude::FromPrimitive, Decimal};

pub struct RuleEngine;

impl RuleEngine {
    pub fn validate_application(
        application: &PolicyApplication,
        product: &Product,
        has_existing_policy: bool,
    ) -> ValidationResult {
        let mut errors = Vec::new();

        if application.insured_amount <= Decimal::ZERO {
            errors.push(ValidationError::InsuredAmountInvalid);
        }

        Self::validate_age(application, product, &mut errors);
        Self::validate_occupation(application, product, &mut errors);
        Self::validate_beneficiaries(&application.beneficiaries, &mut errors);

        if has_existing_policy {
            errors.push(ValidationError::AlreadyInsured);
        }

        let premium = if errors.is_empty() {
            Some(Self::calculate_premium(product, &application.insured_amount))
        } else {
            None
        };

        ValidationResult {
            valid: errors.is_empty(),
            errors,
            premium,
        }
    }

    fn validate_age(
        application: &PolicyApplication,
        product: &Product,
        errors: &mut Vec<ValidationError>,
    ) {
        let age_days = application
            .application_date
            .signed_duration_since(application.insured.birth_date)
            .num_days();

        let min_days = product.product_type.min_age_days();
        let max_days = product.product_type.max_age_days();

        if age_days < min_days || age_days > max_days {
            errors.push(ValidationError::AgeOutOfRange {
                product: product.name.clone(),
                min_age: min_days / 365,
                max_age: max_days / 365,
            });
        }
    }

    fn validate_occupation(
        application: &PolicyApplication,
        product: &Product,
        errors: &mut Vec<ValidationError>,
    ) {
        let risk_level = application.insured.effective_risk_level();

        let allowed = match product.product_type {
            ProductType::Accident => risk_level <= 5,
            ProductType::CriticalIllness => risk_level <= 3,
            ProductType::Life => risk_level <= 3,
        };

        if !allowed {
            errors.push(ValidationError::OccupationNotAllowed {
                level: application.insured.occupation_risk_level,
                product: product.name.clone(),
            });
        }
    }

    fn validate_beneficiaries(beneficiaries: &[Beneficiary], errors: &mut Vec<ValidationError>) {
        if beneficiaries.len() > 3 {
            errors.push(ValidationError::TooManyBeneficiaries {
                count: beneficiaries.len(),
            });
            return;
        }

        let sum: Decimal = beneficiaries.iter().map(|b| b.ratio).sum();
        let hundred = Decimal::from_u8(100).unwrap();

        if sum != hundred {
            errors.push(ValidationError::BeneficiaryRatioSumInvalid { current: sum });
        }
    }

    pub fn calculate_premium(product: &Product, insured_amount: &Decimal) -> Decimal {
        let rate = product.base_premium / Decimal::from_i32(10000).unwrap();
        let premium = insured_amount * rate;
        premium.round_dp(2)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::Days;
    use rust_decimal_macros::dec;

    fn create_test_product(product_type: ProductType) -> Product {
        Product {
            code: match product_type {
                ProductType::Accident => "ACC001".to_string(),
                ProductType::CriticalIllness => "CI001".to_string(),
                ProductType::Life => "LIFE001".to_string(),
            },
            name: match product_type {
                ProductType::Accident => "综合意外险".to_string(),
                ProductType::CriticalIllness => "重大疾病险".to_string(),
                ProductType::Life => "终身寿险".to_string(),
            },
            product_type,
            base_premium: dec!(50),
        }
    }

    #[test]
    fn test_age_validation_exact_day() {
        let product = create_test_product(ProductType::Accident);
        let application_date = NaiveDate::from_ymd_opt(2024, 1, 1).unwrap();

        let birth_date = application_date - Days::new(18 * 365);
        let insured = Insured {
            name: "测试".to_string(),
            id_card: "110101200601011234".to_string(),
            birth_date,
            occupation_risk_level: 1,
            has_safety_training: false,
        };

        let application = PolicyApplication {
            insured: insured.clone(),
            product_code: product.code.clone(),
            application_date,
            beneficiaries: vec![Beneficiary {
                name: "受益人".to_string(),
                id_card: "110101199001011234".to_string(),
                ratio: dec!(100),
            }],
            insured_amount: dec!(100000),
        };

        let result = RuleEngine::validate_application(&application, &product, false);
        assert!(result.valid, "正好满18岁应该可以投保");

        let birth_date = birth_date + Days::new(1);
        let insured = Insured {
            birth_date,
            ..insured
        };
        let application = PolicyApplication {
            insured,
            ..application
        };
        let result = RuleEngine::validate_application(&application, &product, false);
        assert!(!result.valid, "差一天满18岁不能投保");
        assert!(matches!(result.errors[0], ValidationError::AgeOutOfRange { .. }));
    }

    #[test]
    fn test_occupation_risk_level() {
        let accident_product = create_test_product(ProductType::Accident);
        let ci_product = create_test_product(ProductType::CriticalIllness);

        let application_date = NaiveDate::from_ymd_opt(2024, 1, 1).unwrap();
        let birth_date = NaiveDate::from_ymd_opt(1990, 1, 1).unwrap();

        let base_insured = Insured {
            name: "测试".to_string(),
            id_card: "110101199001011234".to_string(),
            birth_date,
            occupation_risk_level: 1,
            has_safety_training: false,
        };

        let base_application = PolicyApplication {
            insured: base_insured,
            product_code: accident_product.code.clone(),
            application_date,
            beneficiaries: vec![Beneficiary {
                name: "受益人".to_string(),
                id_card: "110101199001011234".to_string(),
                ratio: dec!(100),
            }],
            insured_amount: dec!(100000),
        };

        let insured_level4 = Insured {
            occupation_risk_level: 4,
            has_safety_training: false,
            ..base_application.insured.clone()
        };
        let app = PolicyApplication {
            insured: insured_level4,
            ..base_application.clone()
        };
        let result = RuleEngine::validate_application(&app, &ci_product, false);
        assert!(!result.valid, "职业等级4不能投重疾险");

        let insured_level5 = Insured {
            occupation_risk_level: 5,
            has_safety_training: false,
            ..base_application.insured.clone()
        };
        let app = PolicyApplication {
            insured: insured_level5,
            ..base_application.clone()
        };
        let result = RuleEngine::validate_application(&app, &accident_product, false);
        assert!(result.valid, "职业等级5可以投意外险");

        let insured_level4_with_training = Insured {
            occupation_risk_level: 4,
            has_safety_training: true,
            ..base_application.insured.clone()
        };
        let app = PolicyApplication {
            insured: insured_level4_with_training,
            ..base_application.clone()
        };
        let result = RuleEngine::validate_application(&app, &ci_product, false);
        assert!(result.valid, "职业等级4有安全培训证书可降为3级，可以投重疾险");

        let insured_level5_with_training = Insured {
            occupation_risk_level: 5,
            has_safety_training: true,
            ..base_application.insured.clone()
        };
        let app = PolicyApplication {
            insured: insured_level5_with_training,
            ..base_application.clone()
        };
        let result = RuleEngine::validate_application(&app, &ci_product, false);
        assert!(!result.valid, "职业等级5有安全培训证书可降为4级，重疾险要求<=3，这里应该失败");
    }

    #[test]
    fn test_beneficiaries() {
        let product = create_test_product(ProductType::Accident);
        let application_date = NaiveDate::from_ymd_opt(2024, 1, 1).unwrap();
        let birth_date = NaiveDate::from_ymd_opt(1990, 1, 1).unwrap();

        let insured = Insured {
            name: "测试".to_string(),
            id_card: "110101199001011234".to_string(),
            birth_date,
            occupation_risk_level: 1,
            has_safety_training: false,
        };

        let beneficiaries = vec![
            Beneficiary {
                name: "受益人1".to_string(),
                id_card: "110101199001011234".to_string(),
                ratio: dec!(50),
            },
            Beneficiary {
                name: "受益人2".to_string(),
                id_card: "110101199001011235".to_string(),
                ratio: dec!(50),
            },
        ];

        let application = PolicyApplication {
            insured,
            product_code: product.code.clone(),
            application_date,
            beneficiaries,
            insured_amount: dec!(100000),
        };

        let result = RuleEngine::validate_application(&application, &product, false);
        assert!(result.valid, "2个受益人比例之和100%应该通过");

        let beneficiaries = vec![
            Beneficiary {
                name: "受益人1".to_string(),
                id_card: "110101199001011234".to_string(),
                ratio: dec!(30),
            },
            Beneficiary {
                name: "受益人2".to_string(),
                id_card: "110101199001011235".to_string(),
                ratio: dec!(30),
            },
            Beneficiary {
                name: "受益人3".to_string(),
                id_card: "110101199001011236".to_string(),
                ratio: dec!(30),
            },
            Beneficiary {
                name: "受益人4".to_string(),
                id_card: "110101199001011237".to_string(),
                ratio: dec!(10),
            },
        ];
        let application = PolicyApplication {
            beneficiaries,
            ..application
        };
        let result = RuleEngine::validate_application(&application, &product, false);
        assert!(!result.valid, "4个受益人应该失败");
    }

    #[test]
    fn test_beneficiary_ratio_sum() {
        let product = create_test_product(ProductType::Accident);
        let application_date = NaiveDate::from_ymd_opt(2024, 1, 1).unwrap();
        let birth_date = NaiveDate::from_ymd_opt(1990, 1, 1).unwrap();

        let insured = Insured {
            name: "测试".to_string(),
            id_card: "110101199001011234".to_string(),
            birth_date,
            occupation_risk_level: 1,
            has_safety_training: false,
        };

        let beneficiaries = vec![
            Beneficiary {
                name: "受益人1".to_string(),
                id_card: "110101199001011234".to_string(),
                ratio: dec!(60),
            },
            Beneficiary {
                name: "受益人2".to_string(),
                id_card: "110101199001011235".to_string(),
                ratio: dec!(30),
            },
        ];

        let application = PolicyApplication {
            insured,
            product_code: product.code.clone(),
            application_date,
            beneficiaries,
            insured_amount: dec!(100000),
        };

        let result = RuleEngine::validate_application(&application, &product, false);
        assert!(!result.valid, "比例之和90%应该失败");
        assert!(matches!(
            result.errors[0],
            ValidationError::BeneficiaryRatioSumInvalid { .. }
        ));
    }

    #[test]
    fn test_premium_calculation() {
        let product = Product {
            code: "TEST001".to_string(),
            name: "测试产品".to_string(),
            product_type: ProductType::Accident,
            base_premium: dec!(50),
        };

        let insured_amount = dec!(200000);
        let premium = RuleEngine::calculate_premium(&product, &insured_amount);

        assert_eq!(premium, dec!(1000.00), "保费计算应该是 200000 * 0.005 = 1000");
    }
}
