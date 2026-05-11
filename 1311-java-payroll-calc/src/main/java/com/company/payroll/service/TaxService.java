package com.company.payroll.service;

import com.company.payroll.config.TaxConfig;
import com.company.payroll.model.Employee;
import com.company.payroll.model.PaySlip;
import com.company.payroll.model.SpecialDeductionType;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDate;
import java.util.List;

public class TaxService {

    public TaxCalculationResult calculateTax(Employee employee, int year, int month,
                                             BigDecimal currentMonthIncome, BigDecimal currentMonthSocialSecurity,
                                             List<PaySlip> previousPaySlips) {
        
        int monthsFromHire = calculateMonthsFromHire(employee, year, month);
        
        BigDecimal cumulativeIncome = currentMonthIncome;
        BigDecimal cumulativeSocialSecurity = currentMonthSocialSecurity;
        BigDecimal cumulativeTaxWithheld = BigDecimal.ZERO;

        for (PaySlip slip : previousPaySlips) {
            cumulativeIncome = cumulativeIncome.add(slip.getGrossSalary());
            cumulativeSocialSecurity = cumulativeSocialSecurity.add(slip.getTotalSocialSecurityAndHousingFund());
            cumulativeTaxWithheld = cumulativeTaxWithheld.add(slip.getCurrentMonthTax());
        }

        BigDecimal specialDeductionPerMonth = calculateSpecialDeductionPerMonth(employee);
        BigDecimal cumulativeSpecialDeduction = specialDeductionPerMonth.multiply(new BigDecimal(monthsFromHire));
        BigDecimal cumulativeStandardDeduction = TaxConfig.STANDARD_DEDUCTION_PER_MONTH.multiply(new BigDecimal(monthsFromHire));

        BigDecimal cumulativeTaxableIncome = cumulativeIncome
                .subtract(cumulativeSocialSecurity)
                .subtract(cumulativeSpecialDeduction)
                .subtract(cumulativeStandardDeduction);

        if (cumulativeTaxableIncome.compareTo(BigDecimal.ZERO) < 0) {
            cumulativeTaxableIncome = BigDecimal.ZERO;
        }

        TaxConfig.TaxBracket taxBracket = TaxConfig.getTaxBracket(cumulativeTaxableIncome);
        BigDecimal taxRate = taxBracket.getTaxRate();
        BigDecimal quickDeduction = taxBracket.getQuickDeduction();

        BigDecimal cumulativeTax = cumulativeTaxableIncome
                .multiply(taxRate)
                .subtract(quickDeduction)
                .setScale(2, RoundingMode.HALF_UP);

        if (cumulativeTax.compareTo(BigDecimal.ZERO) < 0) {
            cumulativeTax = BigDecimal.ZERO;
        }

        BigDecimal currentMonthTax = cumulativeTax.subtract(cumulativeTaxWithheld);
        if (currentMonthTax.compareTo(BigDecimal.ZERO) < 0) {
            currentMonthTax = BigDecimal.ZERO;
        }

        return new TaxCalculationResult(
                cumulativeIncome,
                cumulativeSocialSecurity,
                cumulativeSpecialDeduction,
                cumulativeStandardDeduction,
                cumulativeTaxableIncome,
                taxRate,
                quickDeduction,
                cumulativeTax,
                cumulativeTaxWithheld,
                currentMonthTax
        );
    }

    public int calculateMonthsFromHire(Employee employee, int year, int month) {
        LocalDate hireDate = employee.getHireDate();
        LocalDate current = LocalDate.of(year, month, 1);

        if (hireDate.isAfter(current)) {
            return 0;
        }

        int hireYear = hireDate.getYear();
        int hireMonth = hireDate.getMonthValue();

        if (year == hireYear) {
            return month - hireMonth + 1;
        } else {
            return month + (12 - hireMonth + 1);
        }
    }

    public BigDecimal calculateSpecialDeductionPerMonth(Employee employee) {
        BigDecimal total = BigDecimal.ZERO;
        for (SpecialDeductionType type : employee.getSpecialDeductions()) {
            total = total.add(type.getMonthlyAmount());
        }
        return total;
    }

    public static class TaxCalculationResult {
        private final BigDecimal cumulativeIncome;
        private final BigDecimal cumulativeSocialSecurity;
        private final BigDecimal cumulativeSpecialDeduction;
        private final BigDecimal cumulativeStandardDeduction;
        private final BigDecimal cumulativeTaxableIncome;
        private final BigDecimal applicableTaxRate;
        private final BigDecimal quickDeduction;
        private final BigDecimal cumulativeTax;
        private final BigDecimal cumulativeTaxWithheld;
        private final BigDecimal currentMonthTax;

        public TaxCalculationResult(BigDecimal cumulativeIncome, BigDecimal cumulativeSocialSecurity,
                                    BigDecimal cumulativeSpecialDeduction, BigDecimal cumulativeStandardDeduction,
                                    BigDecimal cumulativeTaxableIncome, BigDecimal applicableTaxRate,
                                    BigDecimal quickDeduction, BigDecimal cumulativeTax,
                                    BigDecimal cumulativeTaxWithheld, BigDecimal currentMonthTax) {
            this.cumulativeIncome = cumulativeIncome;
            this.cumulativeSocialSecurity = cumulativeSocialSecurity;
            this.cumulativeSpecialDeduction = cumulativeSpecialDeduction;
            this.cumulativeStandardDeduction = cumulativeStandardDeduction;
            this.cumulativeTaxableIncome = cumulativeTaxableIncome;
            this.applicableTaxRate = applicableTaxRate;
            this.quickDeduction = quickDeduction;
            this.cumulativeTax = cumulativeTax;
            this.cumulativeTaxWithheld = cumulativeTaxWithheld;
            this.currentMonthTax = currentMonthTax;
        }

        public BigDecimal getCumulativeIncome() {
            return cumulativeIncome;
        }

        public BigDecimal getCumulativeSocialSecurity() {
            return cumulativeSocialSecurity;
        }

        public BigDecimal getCumulativeSpecialDeduction() {
            return cumulativeSpecialDeduction;
        }

        public BigDecimal getCumulativeStandardDeduction() {
            return cumulativeStandardDeduction;
        }

        public BigDecimal getCumulativeTaxableIncome() {
            return cumulativeTaxableIncome;
        }

        public BigDecimal getApplicableTaxRate() {
            return applicableTaxRate;
        }

        public BigDecimal getQuickDeduction() {
            return quickDeduction;
        }

        public BigDecimal getCumulativeTax() {
            return cumulativeTax;
        }

        public BigDecimal getCumulativeTaxWithheld() {
            return cumulativeTaxWithheld;
        }

        public BigDecimal getCurrentMonthTax() {
            return currentMonthTax;
        }
    }
}
