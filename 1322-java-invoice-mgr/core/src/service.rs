use std::collections::HashMap;
use std::sync::{Arc, RwLock};

use chrono::Local;
use rust_decimal_macros::dec;
use uuid::Uuid;

use crate::config::InvoiceConfig;
use crate::error::InvoiceError;
use crate::models::{
    CreateInvoiceRequest, Invoice, InvoiceStatus, Red冲InvoiceRequest, SearchInvoiceRequest,
    VoidInvoiceRequest,
};

#[derive(Clone)]
pub struct InvoiceService {
    config: InvoiceConfig,
    invoices: Arc<RwLock<HashMap<Uuid, Invoice>>>,
    code_number_index: Arc<RwLock<HashMap<(String, String), Uuid>>>,
}

impl InvoiceService {
    pub fn new(config: InvoiceConfig) -> Self {
        Self {
            config,
            invoices: Arc::new(RwLock::new(HashMap::new())),
            code_number_index: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub fn create_invoice(
        &self,
        req: CreateInvoiceRequest,
    ) -> Result<Invoice, InvoiceError> {
        let today = Local::now().date_naive();
        if req.issue_date > today {
            return Err(InvoiceError::FutureDate);
        }

        if req.amount <= dec!(0) {
            return Err(InvoiceError::AmountMustBePositive);
        }

        let tax_rate = req.tax_rate.unwrap_or(self.config.default_tax_rate);
        if tax_rate < dec!(0) || tax_rate > dec!(1) {
            return Err(InvoiceError::InvalidTaxRate);
        }

        let tax = req.tax.unwrap_or_else(|| {
            let rounded = (req.amount * tax_rate).round_dp(2);
            rounded
        });

        let code_key = (req.invoice_code.clone(), req.invoice_number.clone());
        {
            let index = self.code_number_index.read().unwrap();
            if index.contains_key(&code_key) {
                return Err(InvoiceError::InvoiceNumberExists(
                    req.invoice_code,
                    req.invoice_number,
                ));
            }
        }

        let now = Local::now();
        let id = Uuid::new_v4();

        let invoice = Invoice {
            id,
            invoice_code: req.invoice_code.clone(),
            invoice_number: req.invoice_number.clone(),
            issue_date: req.issue_date,
            buyer_name: req.buyer_name,
            amount: req.amount,
            tax,
            tax_rate,
            status: InvoiceStatus::Normal,
            created_at: now,
            updated_at: now,
            original_invoice_id: None,
            red冲_invoice_id: None,
            is_red冲: false,
        };

        {
            let mut invoices = self.invoices.write().unwrap();
            invoices.insert(id, invoice.clone());
        }

        {
            let mut index = self.code_number_index.write().unwrap();
            index.insert(code_key, id);
        }

        Ok(invoice)
    }

    pub fn get_invoice(&self, id: Uuid) -> Option<Invoice> {
        let invoices = self.invoices.read().unwrap();
        invoices.get(&id).cloned()
    }

    pub fn list_invoices(&self) -> Vec<Invoice> {
        let invoices = self.invoices.read().unwrap();
        let mut list: Vec<Invoice> = invoices.values().cloned().collect();
        list.sort_by(|a, b| a.created_at.cmp(&b.created_at));
        list
    }

    pub fn search_invoices(&self, req: SearchInvoiceRequest) -> Vec<Invoice> {
        let invoices = self.invoices.read().unwrap();
        let mut results: Vec<Invoice> = invoices
            .values()
            .filter(|inv| {
                if let Some(ref buyer) = req.buyer_name {
                    if !inv.buyer_name.to_lowercase().contains(&buyer.to_lowercase()) {
                        return false;
                    }
                }

                if let Some(start) = req.start_date {
                    if inv.issue_date < start {
                        return false;
                    }
                }

                if let Some(end) = req.end_date {
                    if inv.issue_date > end {
                        return false;
                    }
                }

                if let Some(status) = req.status {
                    if inv.status != status {
                        return false;
                    }
                }

                true
            })
            .cloned()
            .collect();

        results.sort_by(|a, b| a.created_at.cmp(&b.created_at));
        results
    }

    pub fn void_invoice(&self, req: VoidInvoiceRequest) -> Result<Invoice, InvoiceError> {
        let mut invoices = self.invoices.write().unwrap();

        let invoice = invoices
            .get_mut(&req.invoice_id)
            .ok_or_else(|| InvoiceError::InvoiceNotFound(req.invoice_id.to_string()))?;

        if invoice.is_red冲 {
            return Err(InvoiceError::Red冲InvoiceCannotBeOperated);
        }

        match invoice.status {
            InvoiceStatus::Normal => {}
            InvoiceStatus::Voided => return Err(InvoiceError::AlreadyVoided),
            InvoiceStatus::Red冲ed => return Err(InvoiceError::AlreadyRed冲ed),
        }

        invoice.status = InvoiceStatus::Voided;
        invoice.updated_at = Local::now();

        Ok(invoice.clone())
    }

    pub fn red冲_invoice(&self, req: Red冲InvoiceRequest) -> Result<Invoice, InvoiceError> {
        let original_invoice = {
            let invoices = self.invoices.read().unwrap();
            invoices
                .get(&req.invoice_id)
                .ok_or_else(|| InvoiceError::InvoiceNotFound(req.invoice_id.to_string()))?
                .clone()
        };

        if original_invoice.is_red冲 {
            return Err(InvoiceError::Red冲InvoiceCannotBeOperated);
        }

        match original_invoice.status {
            InvoiceStatus::Normal => {}
            InvoiceStatus::Voided => return Err(InvoiceError::AlreadyVoided),
            InvoiceStatus::Red冲ed => return Err(InvoiceError::AlreadyRed冲ed),
        }

        let today = Local::now().date_naive();
        let new_invoice_number = format!("{}_R{}", original_invoice.invoice_number, Uuid::new_v4().as_simple().to_string().chars().take(8).collect::<String>());

        let code_key = (original_invoice.invoice_code.clone(), new_invoice_number.clone());
        {
            let index = self.code_number_index.read().unwrap();
            if index.contains_key(&code_key) {
                return Err(InvoiceError::InvoiceNumberExists(
                    original_invoice.invoice_code.clone(),
                    new_invoice_number,
                ));
            }
        }

        let now = Local::now();
        let red冲_id = Uuid::new_v4();

        let red冲_invoice = Invoice {
            id: red冲_id,
            invoice_code: original_invoice.invoice_code.clone(),
            invoice_number: new_invoice_number.clone(),
            issue_date: today,
            buyer_name: original_invoice.buyer_name.clone(),
            amount: -original_invoice.amount,
            tax: -original_invoice.tax,
            tax_rate: original_invoice.tax_rate,
            status: InvoiceStatus::Normal,
            created_at: now,
            updated_at: now,
            original_invoice_id: Some(original_invoice.id),
            red冲_invoice_id: None,
            is_red冲: true,
        };

        {
            let mut invoices = self.invoices.write().unwrap();
            let original = invoices
                .get_mut(&original_invoice.id)
                .ok_or_else(|| InvoiceError::InvoiceNotFound(original_invoice.id.to_string()))?;

            original.status = InvoiceStatus::Red冲ed;
            original.red冲_invoice_id = Some(red冲_id);
            original.updated_at = now;

            invoices.insert(red冲_id, red冲_invoice.clone());
        }

        {
            let mut index = self.code_number_index.write().unwrap();
            index.insert(code_key, red冲_id);
        }

        Ok(red冲_invoice)
    }

    pub fn get_red冲_chain(&self, id: Uuid) -> Option<(Option<Invoice>, Invoice, Option<Invoice>)> {
        let invoice = self.get_invoice(id)?;

        let original = invoice
            .original_invoice_id
            .and_then(|oid| self.get_invoice(oid));

        let red冲 = invoice
            .red冲_invoice_id
            .and_then(|rid| self.get_invoice(rid));

        Some((original, invoice, red冲))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_create_invoice_success() {
        let service = InvoiceService::new(InvoiceConfig::default());
        let today = Local::now().date_naive();

        let req = CreateInvoiceRequest {
            invoice_code: "001".to_string(),
            invoice_number: "0001".to_string(),
            issue_date: today,
            buyer_name: "测试公司".to_string(),
            amount: dec!(1000),
            tax: None,
            tax_rate: None,
        };

        let result = service.create_invoice(req);
        assert!(result.is_ok());

        let invoice = result.unwrap();
        assert_eq!(invoice.amount, dec!(1000));
        assert_eq!(invoice.tax, dec!(130));
        assert_eq!(invoice.status, InvoiceStatus::Normal);
    }

    #[test]
    fn test_duplicate_invoice_number() {
        let service = InvoiceService::new(InvoiceConfig::default());
        let today = Local::now().date_naive();

        let req1 = CreateInvoiceRequest {
            invoice_code: "001".to_string(),
            invoice_number: "0001".to_string(),
            issue_date: today,
            buyer_name: "公司A".to_string(),
            amount: dec!(1000),
            tax: None,
            tax_rate: None,
        };

        let req2 = CreateInvoiceRequest {
            invoice_code: "001".to_string(),
            invoice_number: "0001".to_string(),
            issue_date: today,
            buyer_name: "公司B".to_string(),
            amount: dec!(2000),
            tax: None,
            tax_rate: None,
        };

        assert!(service.create_invoice(req1).is_ok());
        let result2 = service.create_invoice(req2);
        assert!(matches!(result2, Err(InvoiceError::InvoiceNumberExists(_, _))));
    }

    #[test]
    fn test_future_date_rejected() {
        let service = InvoiceService::new(InvoiceConfig::default());
        let future = Local::now().date_naive() + chrono::Duration::days(1);

        let req = CreateInvoiceRequest {
            invoice_code: "001".to_string(),
            invoice_number: "0001".to_string(),
            issue_date: future,
            buyer_name: "测试公司".to_string(),
            amount: dec!(1000),
            tax: None,
            tax_rate: None,
        };

        let result = service.create_invoice(req);
        assert!(matches!(result, Err(InvoiceError::FutureDate)));
    }

    #[test]
    fn test_void_invoice() {
        let service = InvoiceService::new(InvoiceConfig::default());
        let today = Local::now().date_naive();

        let inv = service
            .create_invoice(CreateInvoiceRequest {
                invoice_code: "001".to_string(),
                invoice_number: "0001".to_string(),
                issue_date: today,
                buyer_name: "测试公司".to_string(),
                amount: dec!(1000),
                tax: None,
                tax_rate: None,
            })
            .unwrap();

        let void_result = service.void_invoice(VoidInvoiceRequest { invoice_id: inv.id });
        assert!(void_result.is_ok());
        assert_eq!(void_result.unwrap().status, InvoiceStatus::Voided);

        let void_again = service.void_invoice(VoidInvoiceRequest { invoice_id: inv.id });
        assert!(matches!(void_again, Err(InvoiceError::AlreadyVoided)));
    }

    #[test]
    fn test_red冲_invoice() {
        let service = InvoiceService::new(InvoiceConfig::default());
        let today = Local::now().date_naive();

        let inv = service
            .create_invoice(CreateInvoiceRequest {
                invoice_code: "001".to_string(),
                invoice_number: "0001".to_string(),
                issue_date: today,
                buyer_name: "测试公司".to_string(),
                amount: dec!(1000),
                tax: None,
                tax_rate: None,
            })
            .unwrap();

        let red冲_result = service.red冲_invoice(Red冲InvoiceRequest { invoice_id: inv.id });
        assert!(red冲_result.is_ok());

        let red冲_inv = red冲_result.unwrap();
        assert_eq!(red冲_inv.amount, dec!(-1000));
        assert_eq!(red冲_inv.tax, dec!(-130));
        assert!(red冲_inv.is_red冲);
        assert_eq!(red冲_inv.original_invoice_id, Some(inv.id));

        let original_updated = service.get_invoice(inv.id).unwrap();
        assert_eq!(original_updated.status, InvoiceStatus::Red冲ed);
        assert_eq!(original_updated.red冲_invoice_id, Some(red冲_inv.id));

        let red冲_again = service.red冲_invoice(Red冲InvoiceRequest { invoice_id: inv.id });
        assert!(matches!(red冲_again, Err(InvoiceError::AlreadyRed冲ed)));
    }

    #[test]
    fn test_search_by_buyer_name() {
        let service = InvoiceService::new(InvoiceConfig::default());
        let today = Local::now().date_naive();

        service
            .create_invoice(CreateInvoiceRequest {
                invoice_code: "001".to_string(),
                invoice_number: "0001".to_string(),
                issue_date: today,
                buyer_name: "阿里巴巴".to_string(),
                amount: dec!(1000),
                tax: None,
                tax_rate: None,
            })
            .unwrap();

        service
            .create_invoice(CreateInvoiceRequest {
                invoice_code: "001".to_string(),
                invoice_number: "0002".to_string(),
                issue_date: today,
                buyer_name: "腾讯科技".to_string(),
                amount: dec!(2000),
                tax: None,
                tax_rate: None,
            })
            .unwrap();

        let results = service.search_invoices(SearchInvoiceRequest {
            buyer_name: Some("阿里".to_string()),
            start_date: None,
            end_date: None,
            status: None,
        });

        assert_eq!(results.len(), 1);
        assert_eq!(results[0].buyer_name, "阿里巴巴");
    }
}
