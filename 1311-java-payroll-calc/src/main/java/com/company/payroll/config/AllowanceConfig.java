package com.company.payroll.config;

import java.math.BigDecimal;

public class AllowanceConfig {
    public static final BigDecimal TRANSPORTATION_ALLOWANCE_EXEMPTION = new BigDecimal("500");
    public static final BigDecimal COMMUNICATION_ALLOWANCE_EXEMPTION = new BigDecimal("300");
    public static final BigDecimal MEAL_ALLOWANCE_EXEMPTION = new BigDecimal("1000");

    public static BigDecimal calculateTaxableTransportation(BigDecimal allowance) {
        return calculateTaxableAmount(allowance, TRANSPORTATION_ALLOWANCE_EXEMPTION);
    }

    public static BigDecimal calculateTaxableCommunication(BigDecimal allowance) {
        return calculateTaxableAmount(allowance, COMMUNICATION_ALLOWANCE_EXEMPTION);
    }

    public static BigDecimal calculateTaxableMeal(BigDecimal allowance) {
        return calculateTaxableAmount(allowance, MEAL_ALLOWANCE_EXEMPTION);
    }

    private static BigDecimal calculateTaxableAmount(BigDecimal allowance, BigDecimal exemption) {
        if (allowance == null) {
            return BigDecimal.ZERO;
        }
        return allowance.compareTo(exemption) > 0 ? allowance.subtract(exemption) : BigDecimal.ZERO;
    }
}
