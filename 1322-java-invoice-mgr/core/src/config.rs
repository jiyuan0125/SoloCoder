use rust_decimal::Decimal;
use rust_decimal_macros::dec;

#[derive(Debug, Clone)]
pub struct InvoiceConfig {
    pub default_tax_rate: Decimal,
}

impl Default for InvoiceConfig {
    fn default() -> Self {
        Self {
            default_tax_rate: dec!(0.13),
        }
    }
}

impl InvoiceConfig {
    pub fn new(default_tax_rate: Decimal) -> Result<Self, String> {
        if default_tax_rate < dec!(0) || default_tax_rate > dec!(1) {
            return Err("税率必须在0到1之间".to_string());
        }
        Ok(Self { default_tax_rate })
    }
}
