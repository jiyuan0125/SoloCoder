package com.company.payroll.config;

import java.math.BigDecimal;

public class SocialSecurityConfig {
    public static final BigDecimal BASE_MIN = new BigDecimal("4000");
    public static final BigDecimal BASE_MAX = new BigDecimal("20000");

    public static final BigDecimal PENSION_INSURANCE_RATE = new BigDecimal("0.08");
    public static final BigDecimal MEDICAL_INSURANCE_RATE = new BigDecimal("0.02");
    public static final BigDecimal UNEMPLOYMENT_INSURANCE_RATE = new BigDecimal("0.005");
    public static final BigDecimal HOUSING_FUND_RATE = new BigDecimal("0.07");

    public static BigDecimal calculateBase(BigDecimal actualSalary) {
        if (actualSalary.compareTo(BASE_MIN) < 0) {
            return BASE_MIN;
        }
        if (actualSalary.compareTo(BASE_MAX) > 0) {
            return BASE_MAX;
        }
        return actualSalary;
    }
}
