use std::collections::HashMap;

use crate::models::{Gender, TestResultValue, TestItem, Abnormality};

#[derive(Debug, Clone, PartialEq, Eq, Hash)]
pub struct AgeRange {
    pub min_age: u32,
    pub max_age: u32,
}

impl AgeRange {
    pub fn contains(&self, age: u32) -> bool {
        age >= self.min_age && age <= self.max_age
    }
}

#[derive(Debug, Clone)]
pub struct ReferenceRange {
    pub min: Option<f64>,
    pub max: Option<f64>,
}

impl ReferenceRange {
    pub fn text(&self) -> String {
        match (self.min, self.max) {
            (Some(min), Some(max)) => format!("{}-{}", min, max),
            (Some(min), None) => format!(">{}", min),
            (None, Some(max)) => format!("<{}", max),
            (None, None) => "-".to_string(),
        }
    }
}

#[derive(Debug, Clone)]
pub struct TestReference {
    pub name: String,
    pub ranges: HashMap<(Gender, AgeRange), ReferenceRange>,
    pub unit: Option<String>,
}

impl TestReference {
    pub fn get_reference(&self, gender: Gender, age: u32) -> Option<(&ReferenceRange, String)> {
        for ((ref_gender, age_range), range) in &self.ranges {
            if *ref_gender == gender && age_range.contains(age) {
                return Some((range, self.unit.clone().unwrap_or_default()));
            }
        }
        None
    }

    pub fn get_reference_text(&self, gender: Gender, age: u32) -> Option<String> {
        self.get_reference(gender, age)
            .map(|(range, unit)| {
                if unit.is_empty() {
                    range.text()
                } else {
                    format!("{} {}", range.text(), unit)
                }
            })
    }
}

#[derive(Debug, Clone)]
pub struct ReferenceRegistry {
    tests: HashMap<String, TestReference>,
}

impl ReferenceRegistry {
    pub fn new() -> Self {
        let mut registry = Self {
            tests: HashMap::new(),
        };
        registry.register_standard_tests();
        registry
    }

    fn register_standard_tests(&mut self) {
        self.register_test(self.create_hemoglobin());
        self.register_test(self.create_white_blood_cells());
        self.register_test(self.create_platelets());
        self.register_test(self.create_alt());
        self.register_test(self.create_ast());
        self.register_test(self.create_cholesterol());
        self.register_test(self.create_triglycerides());
        self.register_test(self.create_glucose());
    }

    fn register_test(&mut self, test: TestReference) {
        self.tests.insert(test.name.clone(), test);
    }

    pub fn get_test(&self, name: &str) -> Option<&TestReference> {
        self.tests.get(name)
    }

    fn create_hemoglobin(&self) -> TestReference {
        let mut ranges = HashMap::new();
        ranges.insert(
            (Gender::Male, AgeRange { min_age: 18, max_age: 65 }),
            ReferenceRange { min: Some(120.0), max: Some(160.0) },
        );
        ranges.insert(
            (Gender::Female, AgeRange { min_age: 18, max_age: 65 }),
            ReferenceRange { min: Some(110.0), max: Some(150.0) },
        );
        ranges.insert(
            (Gender::Male, AgeRange { min_age: 0, max_age: 17 }),
            ReferenceRange { min: Some(110.0), max: Some(160.0) },
        );
        ranges.insert(
            (Gender::Female, AgeRange { min_age: 0, max_age: 17 }),
            ReferenceRange { min: Some(110.0), max: Some(150.0) },
        );
        ranges.insert(
            (Gender::Male, AgeRange { min_age: 66, max_age: 150 }),
            ReferenceRange { min: Some(110.0), max: Some(150.0) },
        );
        ranges.insert(
            (Gender::Female, AgeRange { min_age: 66, max_age: 150 }),
            ReferenceRange { min: Some(100.0), max: Some(140.0) },
        );
        TestReference {
            name: "血红蛋白".to_string(),
            ranges,
            unit: Some("g/L".to_string()),
        }
    }

    fn create_white_blood_cells(&self) -> TestReference {
        let mut ranges = HashMap::new();
        ranges.insert(
            (Gender::Male, AgeRange { min_age: 0, max_age: 150 }),
            ReferenceRange { min: Some(4.0), max: Some(10.0) },
        );
        ranges.insert(
            (Gender::Female, AgeRange { min_age: 0, max_age: 150 }),
            ReferenceRange { min: Some(4.0), max: Some(10.0) },
        );
        TestReference {
            name: "白细胞计数".to_string(),
            ranges,
            unit: Some("×10^9/L".to_string()),
        }
    }

    fn create_platelets(&self) -> TestReference {
        let mut ranges = HashMap::new();
        ranges.insert(
            (Gender::Male, AgeRange { min_age: 0, max_age: 150 }),
            ReferenceRange { min: Some(100.0), max: Some(300.0) },
        );
        ranges.insert(
            (Gender::Female, AgeRange { min_age: 0, max_age: 150 }),
            ReferenceRange { min: Some(100.0), max: Some(300.0) },
        );
        TestReference {
            name: "血小板计数".to_string(),
            ranges,
            unit: Some("×10^9/L".to_string()),
        }
    }

    fn create_alt(&self) -> TestReference {
        let mut ranges = HashMap::new();
        ranges.insert(
            (Gender::Male, AgeRange { min_age: 18, max_age: 150 }),
            ReferenceRange { min: Some(0.0), max: Some(40.0) },
        );
        ranges.insert(
            (Gender::Female, AgeRange { min_age: 18, max_age: 150 }),
            ReferenceRange { min: Some(0.0), max: Some(35.0) },
        );
        ranges.insert(
            (Gender::Male, AgeRange { min_age: 0, max_age: 17 }),
            ReferenceRange { min: Some(0.0), max: Some(30.0) },
        );
        ranges.insert(
            (Gender::Female, AgeRange { min_age: 0, max_age: 17 }),
            ReferenceRange { min: Some(0.0), max: Some(25.0) },
        );
        TestReference {
            name: "谷丙转氨酶(ALT)".to_string(),
            ranges,
            unit: Some("U/L".to_string()),
        }
    }

    fn create_ast(&self) -> TestReference {
        let mut ranges = HashMap::new();
        ranges.insert(
            (Gender::Male, AgeRange { min_age: 18, max_age: 150 }),
            ReferenceRange { min: Some(0.0), max: Some(40.0) },
        );
        ranges.insert(
            (Gender::Female, AgeRange { min_age: 18, max_age: 150 }),
            ReferenceRange { min: Some(0.0), max: Some(35.0) },
        );
        TestReference {
            name: "谷草转氨酶(AST)".to_string(),
            ranges,
            unit: Some("U/L".to_string()),
        }
    }

    fn create_cholesterol(&self) -> TestReference {
        let mut ranges = HashMap::new();
        ranges.insert(
            (Gender::Male, AgeRange { min_age: 0, max_age: 150 }),
            ReferenceRange { min: Some(2.8), max: Some(5.2) },
        );
        ranges.insert(
            (Gender::Female, AgeRange { min_age: 0, max_age: 150 }),
            ReferenceRange { min: Some(2.8), max: Some(5.2) },
        );
        TestReference {
            name: "总胆固醇".to_string(),
            ranges,
            unit: Some("mmol/L".to_string()),
        }
    }

    fn create_triglycerides(&self) -> TestReference {
        let mut ranges = HashMap::new();
        ranges.insert(
            (Gender::Male, AgeRange { min_age: 0, max_age: 150 }),
            ReferenceRange { min: Some(0.4), max: Some(1.7) },
        );
        ranges.insert(
            (Gender::Female, AgeRange { min_age: 0, max_age: 150 }),
            ReferenceRange { min: Some(0.4), max: Some(1.7) },
        );
        TestReference {
            name: "甘油三酯".to_string(),
            ranges,
            unit: Some("mmol/L".to_string()),
        }
    }

    fn create_glucose(&self) -> TestReference {
        let mut ranges = HashMap::new();
        ranges.insert(
            (Gender::Male, AgeRange { min_age: 0, max_age: 150 }),
            ReferenceRange { min: Some(3.9), max: Some(6.1) },
        );
        ranges.insert(
            (Gender::Female, AgeRange { min_age: 0, max_age: 150 }),
            ReferenceRange { min: Some(3.9), max: Some(6.1) },
        );
        TestReference {
            name: "空腹血糖".to_string(),
            ranges,
            unit: Some("mmol/L".to_string()),
        }
    }
}

pub fn evaluate_test(
    test_name: &str,
    value: &TestResultValue,
    unit: Option<&str>,
    gender: Gender,
    age: u32,
    registry: &ReferenceRegistry,
) -> TestItem {
    let reference_text = registry.get_test(test_name).and_then(|t| t.get_reference_text(gender, age));
    
    let abnormality = match value {
        TestResultValue::Quantitative(val) => {
            if let Some(test_ref) = registry.get_test(test_name) {
                if let Some((range, _)) = test_ref.get_reference(gender, age) {
                    if let Some(min) = range.min {
                        if *val < min {
                            return TestItem {
                                name: test_name.to_string(),
                                value: value.clone(),
                                unit: unit.map(|u| u.to_string()),
                                reference_text,
                                abnormality: Some(Abnormality::Low),
                            };
                        }
                    }
                    if let Some(max) = range.max {
                        if *val > max {
                            return TestItem {
                                name: test_name.to_string(),
                                value: value.clone(),
                                unit: unit.map(|u| u.to_string()),
                                reference_text,
                                abnormality: Some(Abnormality::High),
                            };
                        }
                    }
                }
            }
            None
        }
        TestResultValue::Qualitative(_) => None,
    };
    
    TestItem {
        name: test_name.to_string(),
        value: value.clone(),
        unit: unit.map(|u| u.to_string()),
        reference_text,
        abnormality,
    }
}
