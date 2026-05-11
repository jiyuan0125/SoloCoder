use std::collections::HashMap;
use uuid::Uuid;
use chrono::NaiveDate;

use crate::models::*;
use crate::error::*;

pub struct CheckupService {
    items: HashMap<Uuid, CheckupItem>,
    packages: HashMap<Uuid, CheckupPackage>,
    customers: HashMap<Uuid, Customer>,
    mutex_rules: Vec<MutexRule>,
    reservations: HashMap<Uuid, Reservation>,
}

impl CheckupService {
    pub fn new() -> Self {
        Self {
            items: HashMap::new(),
            packages: HashMap::new(),
            customers: HashMap::new(),
            mutex_rules: Vec::new(),
            reservations: HashMap::new(),
        }
    }

    pub fn add_item(&mut self, item: CheckupItem) {
        self.items.insert(item.id, item);
    }

    pub fn get_item(&self, id: &Uuid) -> Option<&CheckupItem> {
        self.items.get(id)
    }

    pub fn list_items(&self) -> Vec<&CheckupItem> {
        self.items.values().collect()
    }

    pub fn add_package(&mut self, package: CheckupPackage) {
        self.packages.insert(package.id, package);
    }

    pub fn get_package(&self, id: &Uuid) -> Option<&CheckupPackage> {
        self.packages.get(id)
    }

    pub fn list_packages(&self) -> Vec<&CheckupPackage> {
        self.packages.values().collect()
    }

    pub fn add_customer(&mut self, customer: Customer) {
        self.customers.insert(customer.id, customer);
    }

    pub fn get_customer(&self, id: &Uuid) -> Option<&Customer> {
        self.customers.get(id)
    }

    pub fn list_customers(&self) -> Vec<&Customer> {
        self.customers.values().collect()
    }

    pub fn add_mutex_rule(&mut self, rule: MutexRule) {
        self.mutex_rules.push(rule);
    }

    pub fn list_mutex_rules(&self) -> &Vec<MutexRule> {
        &self.mutex_rules
    }

    pub fn remove_mutex_rule(&mut self, rule_id: &Uuid) {
        self.mutex_rules.retain(|r| r.id != *rule_id);
    }

    pub fn create_reservation(
        &mut self,
        customer_id: Uuid,
        appointment_date: NaiveDate,
        base_package: Option<Uuid>,
        additional_items: Vec<Uuid>,
        removed_items: Vec<Uuid>,
    ) -> Result<Reservation, CheckupError> {
        let customer = self.customers.get(&customer_id)
            .ok_or_else(|| CheckupError::CustomerNotFound(format!("{}", customer_id)))?;

        let age = customer.calculate_age(appointment_date);

        let mut all_items: Vec<Uuid> = Vec::new();
        
        if let Some(pkg_id) = base_package {
            let pkg = self.packages.get(&pkg_id)
                .ok_or_else(|| CheckupError::PackageNotFound(format!("{}", pkg_id)))?;
            all_items.extend(pkg.items.iter().cloned());
        }

        all_items.extend(additional_items.iter().cloned());

        all_items.retain(|item_id| !removed_items.contains(item_id));

        self.validate_selection(&customer, age, &all_items)?;

        let (warnings, requires_approval) = self.check_warnings(&customer, age, &all_items);

        let (total_price, refund_amount) = self.calculate_pricing(
            base_package,
            &additional_items,
            &removed_items,
        )?;

        let reservation = Reservation {
            id: Uuid::new_v4(),
            customer_id,
            appointment_date,
            base_package,
            selected_items: all_items,
            removed_items,
            warnings,
            requires_doctor_approval: requires_approval,
            total_price,
            refund_amount,
        };

        self.reservations.insert(reservation.id, reservation.clone());

        Ok(reservation)
    }

    pub fn get_reservation(&self, id: &Uuid) -> Option<&Reservation> {
        self.reservations.get(id)
    }

    pub fn list_reservations(&self) -> Vec<&Reservation> {
        self.reservations.values().collect()
    }

    fn validate_selection(
        &self,
        customer: &Customer,
        age: i32,
        items: &[Uuid],
    ) -> Result<(), CheckupError> {
        for item_id in items {
            let item = self.items.get(item_id)
                .ok_or_else(|| CheckupError::ItemNotFound(format!("{}", item_id)))?;

            self.check_gender_restriction(customer, item)?;
            self.check_pregnant_restriction(customer, item)?;
            self.check_age_restriction(age, item)?;
        }

        self.check_hard_mutex_rules(items)?;

        Ok(())
    }

    fn check_gender_restriction(
        &self,
        customer: &Customer,
        item: &CheckupItem,
    ) -> Result<(), CheckupError> {
        match item.gender_restriction {
            GenderRestriction::MaleOnly => {
                if customer.gender != Gender::Male {
                    return Err(CheckupError::GenderRestrictionViolation(
                        format!("项目 {} 仅限男性", item.name)
                    ));
                }
            }
            GenderRestriction::FemaleOnly => {
                if customer.gender != Gender::Female {
                    return Err(CheckupError::GenderRestrictionViolation(
                        format!("项目 {} 仅限女性", item.name)
                    ));
                }
            }
            GenderRestriction::None => {}
        }
        Ok(())
    }

    fn check_pregnant_restriction(
        &self,
        customer: &Customer,
        item: &CheckupItem,
    ) -> Result<(), CheckupError> {
        if customer.is_pregnant && item.is_pregnant_forbidden {
            return Err(CheckupError::PregnantForbidden(
                format!("项目 {} 孕妇禁忌", item.name)
            ));
        }
        Ok(())
    }

    fn check_age_restriction(
        &self,
        age: i32,
        item: &CheckupItem,
    ) -> Result<(), CheckupError> {
        if age >= 80 && item.has_contrast_agent {
            return Err(CheckupError::AgeRestrictionViolation(
                format!("80岁以上不建议做注射造影剂项目: {}", item.name)
            ));
        }
        Ok(())
    }

    fn check_hard_mutex_rules(
        &self,
        items: &[Uuid],
    ) -> Result<(), CheckupError> {
        for rule in &self.mutex_rules {
            if rule.rule_type == MutexRuleType::Hard {
                let has_item1 = items.contains(&rule.item1);
                let has_item2 = items.contains(&rule.item2);
                
                if has_item1 && has_item2 {
                    let item1 = self.items.get(&rule.item1).map(|i| i.name.as_str()).unwrap_or("未知");
                    let item2 = self.items.get(&rule.item2).map(|i| i.name.as_str()).unwrap_or("未知");
                    return Err(CheckupError::HardMutexConflict(
                        format!("{} 与 {} 互斥: {}", item1, item2, rule.description)
                    ));
                }
            }
        }
        Ok(())
    }

    fn check_warnings(
        &self,
        _customer: &Customer,
        age: i32,
        items: &[Uuid],
    ) -> (Vec<String>, Vec<Uuid>) {
        let mut warnings = Vec::new();
        let mut requires_approval = Vec::new();

        for item_id in items {
            if let Some(item) = self.items.get(item_id) {
                if (item.is_gastroscopy || item.is_colonoscopy) && age >= 60 {
                    warnings.push(format!(
                        "{}：60岁以上做胃肠镜需要医生额外确认",
                        item.name
                    ));
                    requires_approval.push(*item_id);
                }
            }
        }

        for rule in &self.mutex_rules {
            if rule.rule_type == MutexRuleType::Warning {
                let has_item1 = items.contains(&rule.item1);
                let has_item2 = items.contains(&rule.item2);
                
                if has_item1 && has_item2 {
                    warnings.push(rule.description.clone());
                }
            }
        }

        (warnings, requires_approval)
    }

    fn calculate_pricing(
        &self,
        base_package: Option<Uuid>,
        additional_items: &[Uuid],
        removed_items: &[Uuid],
    ) -> Result<(f64, f64), CheckupError> {
        let mut total = 0.0;
        let mut refund = 0.0;

        if let Some(pkg_id) = base_package {
            let pkg = self.packages.get(&pkg_id)
                .ok_or_else(|| CheckupError::PackageNotFound(format!("{}", pkg_id)))?;
            total += pkg.package_price;
        }

        for item_id in additional_items {
            let item = self.items.get(item_id)
                .ok_or_else(|| CheckupError::ItemNotFound(format!("{}", item_id)))?;
            total += item.price;
        }

        for item_id in removed_items {
            let item = self.items.get(item_id)
                .ok_or_else(|| CheckupError::ItemNotFound(format!("{}", item_id)))?;
            refund += item.price * 0.8;
        }

        total -= refund;

        Ok((total, refund))
    }
}
