package com.company.payroll.config;

import java.math.BigDecimal;
import java.util.ArrayList;
import java.util.List;

public class TaxConfig {
    public static final BigDecimal STANDARD_DEDUCTION_PER_MONTH = new BigDecimal("5000");

    public static final List<TaxBracket> TAX_BRACKETS = new ArrayList<>();

    static {
        TAX_BRACKETS.add(new TaxBracket(BigDecimal.ZERO, new BigDecimal("36000"), new BigDecimal("0.03"), BigDecimal.ZERO));
        TAX_BRACKETS.add(new TaxBracket(new BigDecimal("36000"), new BigDecimal("144000"), new BigDecimal("0.10"), new BigDecimal("2520")));
        TAX_BRACKETS.add(new TaxBracket(new BigDecimal("144000"), new BigDecimal("300000"), new BigDecimal("0.20"), new BigDecimal("16920")));
        TAX_BRACKETS.add(new TaxBracket(new BigDecimal("300000"), new BigDecimal("420000"), new BigDecimal("0.25"), new BigDecimal("31920")));
        TAX_BRACKETS.add(new TaxBracket(new BigDecimal("420000"), new BigDecimal("660000"), new BigDecimal("0.30"), new BigDecimal("52920")));
        TAX_BRACKETS.add(new TaxBracket(new BigDecimal("660000"), new BigDecimal("960000"), new BigDecimal("0.35"), new BigDecimal("85920")));
        TAX_BRACKETS.add(new TaxBracket(new BigDecimal("960000"), null, new BigDecimal("0.45"), new BigDecimal("181920")));
    }

    public static TaxBracket getTaxBracket(BigDecimal taxableIncome) {
        for (int i = TAX_BRACKETS.size() - 1; i >= 0; i--) {
            TaxBracket bracket = TAX_BRACKETS.get(i);
            if (taxableIncome.compareTo(bracket.getLowerBound()) >= 0) {
                if (bracket.getUpperBound() == null || taxableIncome.compareTo(bracket.getUpperBound()) < 0) {
                    return bracket;
                }
            }
        }
        return TAX_BRACKETS.get(0);
    }

    public static class TaxBracket {
        private final BigDecimal lowerBound;
        private final BigDecimal upperBound;
        private final BigDecimal taxRate;
        private final BigDecimal quickDeduction;

        public TaxBracket(BigDecimal lowerBound, BigDecimal upperBound, BigDecimal taxRate, BigDecimal quickDeduction) {
            this.lowerBound = lowerBound;
            this.upperBound = upperBound;
            this.taxRate = taxRate;
            this.quickDeduction = quickDeduction;
        }

        public BigDecimal getLowerBound() {
            return lowerBound;
        }

        public BigDecimal getUpperBound() {
            return upperBound;
        }

        public BigDecimal getTaxRate() {
            return taxRate;
        }

        public BigDecimal getQuickDeduction() {
            return quickDeduction;
        }
    }
}
