package com.company.payroll.service;

import com.company.payroll.config.TaxConfig;
import com.company.payroll.model.Employee;
import com.company.payroll.model.PaySlip;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.util.Arrays;
import java.util.List;

public class YearEndBonusService {

    public YearEndBonusResult calculateBonus(BigDecimal bonusAmount, Employee employee,
                                             int year, int month, BigDecimal currentMonthIncome,
                                             BigDecimal currentMonthSocialSecurity,
                                             List<PaySlip> previousPaySlips) {

        if (bonusAmount == null || bonusAmount.compareTo(BigDecimal.ZERO) <= 0) {
            return new YearEndBonusResult(BigDecimal.ZERO, BigDecimal.ZERO, BigDecimal.ZERO, BigDecimal.ZERO);
        }

        BigDecimal taxSeparately = calculateSeparatelyTax(bonusAmount);

        BigDecimal taxCombined = calculateCombinedTax(bonusAmount, employee, year, month,
                currentMonthIncome, currentMonthSocialSecurity, previousPaySlips);

        BigDecimal netBonusSeparately = bonusAmount.subtract(taxSeparately);
        BigDecimal netBonusCombined = bonusAmount.subtract(taxCombined);

        return new YearEndBonusResult(bonusAmount, taxSeparately, taxCombined,
                taxSeparately.compareTo(taxCombined) <= 0 ? netBonusSeparately : netBonusCombined);
    }

    private BigDecimal calculateSeparatelyTax(BigDecimal bonusAmount) {
        BigDecimal monthlyAmount = bonusAmount.divide(new BigDecimal("12"), 10, RoundingMode.DOWN);
        BigDecimal roundedMonthly = monthlyAmount.setScale(0, RoundingMode.HALF_UP);

        MonthTaxBracket bracket = findMonthTaxBracket(roundedMonthly);

        BigDecimal tax = bonusAmount.multiply(bracket.taxRate)
                .subtract(bracket.quickDeduction)
                .setScale(2, RoundingMode.HALF_UP);

        if (tax.compareTo(BigDecimal.ZERO) < 0) {
            return BigDecimal.ZERO;
        }
        return tax;
    }

    private static final List<MonthTaxBracket> MONTH_TAX_BRACKETS = Arrays.asList(
            new MonthTaxBracket(BigDecimal.ZERO, new BigDecimal("3000"), new BigDecimal("0.03"), BigDecimal.ZERO),
            new MonthTaxBracket(new BigDecimal("3000"), new BigDecimal("12000"), new BigDecimal("0.10"), new BigDecimal("210")),
            new MonthTaxBracket(new BigDecimal("12000"), new BigDecimal("25000"), new BigDecimal("0.20"), new BigDecimal("1410")),
            new MonthTaxBracket(new BigDecimal("25000"), new BigDecimal("35000"), new BigDecimal("0.25"), new BigDecimal("2660")),
            new MonthTaxBracket(new BigDecimal("35000"), new BigDecimal("55000"), new BigDecimal("0.30"), new BigDecimal("4410")),
            new MonthTaxBracket(new BigDecimal("55000"), new BigDecimal("80000"), new BigDecimal("0.35"), new BigDecimal("7160")),
            new MonthTaxBracket(new BigDecimal("80000"), null, new BigDecimal("0.45"), new BigDecimal("15160"))
    );

    private MonthTaxBracket findMonthTaxBracket(BigDecimal monthlyAmount) {
        for (int i = MONTH_TAX_BRACKETS.size() - 1; i >= 0; i--) {
            MonthTaxBracket bracket = MONTH_TAX_BRACKETS.get(i);
            if (monthlyAmount.compareTo(bracket.lowerBound) >= 0) {
                if (bracket.upperBound == null || monthlyAmount.compareTo(bracket.upperBound) < 0) {
                    return bracket;
                }
            }
        }
        return MONTH_TAX_BRACKETS.get(0);
    }

    private static class MonthTaxBracket {
        final BigDecimal lowerBound;
        final BigDecimal upperBound;
        final BigDecimal taxRate;
        final BigDecimal quickDeduction;

        MonthTaxBracket(BigDecimal lowerBound, BigDecimal upperBound, BigDecimal taxRate, BigDecimal quickDeduction) {
            this.lowerBound = lowerBound;
            this.upperBound = upperBound;
            this.taxRate = taxRate;
            this.quickDeduction = quickDeduction;
        }
    }

    private BigDecimal calculateCombinedTax(BigDecimal bonusAmount, Employee employee,
                                            int year, int month, BigDecimal currentMonthIncome,
                                            BigDecimal currentMonthSocialSecurity,
                                            List<PaySlip> previousPaySlips) {

        TaxService taxService = new TaxService();
        
        BigDecimal incomeWithoutBonus = currentMonthIncome;
        TaxService.TaxCalculationResult resultWithoutBonus = taxService.calculateTax(
                employee, year, month, incomeWithoutBonus, currentMonthSocialSecurity, previousPaySlips);
        BigDecimal taxWithoutBonus = resultWithoutBonus.getCurrentMonthTax();

        BigDecimal incomeWithBonus = currentMonthIncome.add(bonusAmount);
        TaxService.TaxCalculationResult resultWithBonus = taxService.calculateTax(
                employee, year, month, incomeWithBonus, currentMonthSocialSecurity, previousPaySlips);
        BigDecimal taxWithBonus = resultWithBonus.getCurrentMonthTax();

        return taxWithBonus.subtract(taxWithoutBonus);
    }

    public static class YearEndBonusResult {
        private final BigDecimal bonusAmount;
        private final BigDecimal taxSeparately;
        private final BigDecimal taxCombined;
        private final BigDecimal netBonus;

        public YearEndBonusResult(BigDecimal bonusAmount, BigDecimal taxSeparately,
                                  BigDecimal taxCombined, BigDecimal netBonus) {
            this.bonusAmount = bonusAmount;
            this.taxSeparately = taxSeparately;
            this.taxCombined = taxCombined;
            this.netBonus = netBonus;
        }

        public BigDecimal getBonusAmount() {
            return bonusAmount;
        }

        public BigDecimal getTaxSeparately() {
            return taxSeparately;
        }

        public BigDecimal getTaxCombined() {
            return taxCombined;
        }

        public BigDecimal getNetBonus() {
            return netBonus;
        }

        public boolean isSeparatelyBetter() {
            return taxSeparately.compareTo(taxCombined) <= 0;
        }
    }
}
