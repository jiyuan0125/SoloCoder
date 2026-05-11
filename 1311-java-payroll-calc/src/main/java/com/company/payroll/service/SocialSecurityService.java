package com.company.payroll.service;

import com.company.payroll.config.SocialSecurityConfig;
import com.company.payroll.model.Employee;

import java.math.BigDecimal;
import java.math.RoundingMode;

public class SocialSecurityService {

    public SocialSecurityResult calculate(Employee employee, int year, int month) {
        BigDecimal baseSalary = employee.getSocialSecurityBaseSalary(year, month);
        BigDecimal base = SocialSecurityConfig.calculateBase(baseSalary);

        BigDecimal pensionInsurance = base.multiply(SocialSecurityConfig.PENSION_INSURANCE_RATE)
                .setScale(2, RoundingMode.HALF_UP);
        BigDecimal medicalInsurance = base.multiply(SocialSecurityConfig.MEDICAL_INSURANCE_RATE)
                .setScale(2, RoundingMode.HALF_UP);
        BigDecimal unemploymentInsurance = base.multiply(SocialSecurityConfig.UNEMPLOYMENT_INSURANCE_RATE)
                .setScale(2, RoundingMode.HALF_UP);
        BigDecimal housingFund = base.multiply(SocialSecurityConfig.HOUSING_FUND_RATE)
                .setScale(2, RoundingMode.HALF_UP);

        BigDecimal total = pensionInsurance
                .add(medicalInsurance)
                .add(unemploymentInsurance)
                .add(housingFund);

        return new SocialSecurityResult(base, pensionInsurance, medicalInsurance, 
                unemploymentInsurance, housingFund, total);
    }

    public static class SocialSecurityResult {
        private final BigDecimal base;
        private final BigDecimal pensionInsurance;
        private final BigDecimal medicalInsurance;
        private final BigDecimal unemploymentInsurance;
        private final BigDecimal housingFund;
        private final BigDecimal total;

        public SocialSecurityResult(BigDecimal base, BigDecimal pensionInsurance, 
                                    BigDecimal medicalInsurance, BigDecimal unemploymentInsurance,
                                    BigDecimal housingFund, BigDecimal total) {
            this.base = base;
            this.pensionInsurance = pensionInsurance;
            this.medicalInsurance = medicalInsurance;
            this.unemploymentInsurance = unemploymentInsurance;
            this.housingFund = housingFund;
            this.total = total;
        }

        public BigDecimal getBase() {
            return base;
        }

        public BigDecimal getPensionInsurance() {
            return pensionInsurance;
        }

        public BigDecimal getMedicalInsurance() {
            return medicalInsurance;
        }

        public BigDecimal getUnemploymentInsurance() {
            return unemploymentInsurance;
        }

        public BigDecimal getHousingFund() {
            return housingFund;
        }

        public BigDecimal getTotal() {
            return total;
        }
    }
}
