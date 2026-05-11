package com.purchase.approval.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;

@Data
@Component
@ConfigurationProperties(prefix = "purchase.rules")
public class PurchaseRulesConfig {
    private BigDecimal skipComparisonAmount = new BigDecimal("5000");
    private BigDecimal twoSupplierMinAmount = new BigDecimal("50000");

    public boolean canSkipComparison(BigDecimal amount) {
        return amount.compareTo(skipComparisonAmount) <= 0;
    }

    public boolean needsTwoSuppliers(BigDecimal amount) {
        return amount.compareTo(skipComparisonAmount) > 0 && amount.compareTo(twoSupplierMinAmount) <= 0;
    }

    public boolean needsThreeSuppliers(BigDecimal amount) {
        return amount.compareTo(twoSupplierMinAmount) > 0;
    }

    public int getRequiredSupplierCount(BigDecimal amount) {
        if (canSkipComparison(amount)) {
            return 1;
        } else if (needsTwoSuppliers(amount)) {
            return 2;
        } else {
            return 3;
        }
    }

    public boolean needsPublicInquiry(BigDecimal amount) {
        return amount.compareTo(twoSupplierMinAmount) > 0;
    }
}
