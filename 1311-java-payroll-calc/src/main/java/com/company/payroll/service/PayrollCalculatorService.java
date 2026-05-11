package com.company.payroll.service;

import com.company.payroll.config.AllowanceConfig;
import com.company.payroll.model.*;
import com.company.payroll.service.SocialSecurityService.SocialSecurityResult;
import com.company.payroll.service.TaxService.TaxCalculationResult;
import com.company.payroll.service.YearEndBonusService.YearEndBonusResult;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDate;
import java.util.List;
import java.util.UUID;

public class PayrollCalculatorService {
    private final PayrollDataStore dataStore;
    private final SocialSecurityService socialSecurityService;
    private final TaxService taxService;
    private final YearEndBonusService yearEndBonusService;

    public PayrollCalculatorService(PayrollDataStore dataStore) {
        this.dataStore = dataStore;
        this.socialSecurityService = new SocialSecurityService();
        this.taxService = new TaxService();
        this.yearEndBonusService = new YearEndBonusService();
    }

    public PaySlip calculatePaySlip(SalaryInput input) {
        Employee employee = dataStore.getEmployee(input.getEmployeeId());
        if (employee == null) {
            throw new IllegalArgumentException("员工不存在: " + input.getEmployeeId());
        }

        int year = input.getYear();
        int month = input.getMonth();

        if (!employee.isActiveAt(year, month)) {
            throw new IllegalArgumentException("员工在该月份不活跃");
        }

        BigDecimal baseSalary = calculateProratedSalary(employee, input);

        BigDecimal performanceBonus = getAdditionalIncome(input, IncomeType.PERFORMANCE_BONUS);
        BigDecimal overtimePay = getAdditionalIncome(input, IncomeType.OVERTIME_PAY);
        BigDecimal projectBonus = getAdditionalIncome(input, IncomeType.PROJECT_BONUS);
        BigDecimal partTimeIncome = getAdditionalIncome(input, IncomeType.PART_TIME_INCOME);
        BigDecimal transportationAllowance = getAdditionalIncome(input, IncomeType.TRANSPORTATION_ALLOWANCE);
        BigDecimal communicationAllowance = getAdditionalIncome(input, IncomeType.COMMUNICATION_ALLOWANCE);
        BigDecimal mealAllowance = getAdditionalIncome(input, IncomeType.MEAL_ALLOWANCE);

        BigDecimal taxableTransportation = AllowanceConfig.calculateTaxableTransportation(transportationAllowance);
        BigDecimal taxableCommunication = AllowanceConfig.calculateTaxableCommunication(communicationAllowance);
        BigDecimal taxableMeal = AllowanceConfig.calculateTaxableMeal(mealAllowance);
        BigDecimal taxableAllowance = taxableTransportation.add(taxableCommunication).add(taxableMeal);

        BigDecimal nonAllowanceIncome = baseSalary
                .add(performanceBonus)
                .add(overtimePay)
                .add(projectBonus)
                .add(partTimeIncome);

        BigDecimal grossSalary = nonAllowanceIncome
                .add(transportationAllowance)
                .add(communicationAllowance)
                .add(mealAllowance);

        BigDecimal taxableIncome = nonAllowanceIncome.add(taxableAllowance);

        SocialSecurityResult socialSecurity = socialSecurityService.calculate(employee, year, month);

        List<PaySlip> previousPaySlips = dataStore.getPreviousPaySlips(employee.getId(), year, month);

        TaxCalculationResult taxResult = taxService.calculateTax(
                employee, year, month, taxableIncome, socialSecurity.getTotal(), previousPaySlips);

        BigDecimal yearEndBonus = input.getYearEndBonus();
        BigDecimal yearEndBonusTaxSeparately = BigDecimal.ZERO;
        BigDecimal yearEndBonusTaxCombined = BigDecimal.ZERO;

        if (yearEndBonus.compareTo(BigDecimal.ZERO) > 0) {
            YearEndBonusResult bonusResult = yearEndBonusService.calculateBonus(
                    yearEndBonus, employee, year, month, taxableIncome,
                    socialSecurity.getTotal(), previousPaySlips);
            yearEndBonusTaxSeparately = bonusResult.getTaxSeparately();
            yearEndBonusTaxCombined = bonusResult.getTaxCombined();
        }

        BigDecimal specialDeductionTotal = taxService.calculateSpecialDeductionPerMonth(employee);

        BigDecimal netSalary = grossSalary
                .subtract(socialSecurity.getTotal())
                .subtract(taxResult.getCurrentMonthTax());

        if (yearEndBonus.compareTo(BigDecimal.ZERO) > 0) {
            BigDecimal bonusTax = yearEndBonusTaxSeparately.compareTo(yearEndBonusTaxCombined) <= 0
                    ? yearEndBonusTaxSeparately : yearEndBonusTaxCombined;
            netSalary = netSalary.add(yearEndBonus).subtract(bonusTax);
        }

        PaySlip paySlip = new PaySlip();
        paySlip.setId(UUID.randomUUID().toString());
        paySlip.setEmployeeId(employee.getId());
        paySlip.setEmployeeName(employee.getName());
        paySlip.setYear(year);
        paySlip.setMonth(month);

        paySlip.setGrossSalary(grossSalary);
        paySlip.setBaseSalary(baseSalary);
        paySlip.setPerformanceBonus(performanceBonus);
        paySlip.setOvertimePay(overtimePay);
        paySlip.setProjectBonus(projectBonus);
        paySlip.setPartTimeIncome(partTimeIncome);
        paySlip.setTransportationAllowance(transportationAllowance);
        paySlip.setCommunicationAllowance(communicationAllowance);
        paySlip.setMealAllowance(mealAllowance);
        paySlip.setTaxableAllowance(taxableAllowance);

        paySlip.setPensionInsurance(socialSecurity.getPensionInsurance());
        paySlip.setMedicalInsurance(socialSecurity.getMedicalInsurance());
        paySlip.setUnemploymentInsurance(socialSecurity.getUnemploymentInsurance());
        paySlip.setHousingFund(socialSecurity.getHousingFund());
        paySlip.setTotalSocialSecurityAndHousingFund(socialSecurity.getTotal());

        paySlip.setSpecialDeductionTotal(specialDeductionTotal);

        paySlip.setCumulativeIncome(taxResult.getCumulativeIncome());
        paySlip.setCumulativeSocialSecurity(taxResult.getCumulativeSocialSecurity());
        paySlip.setCumulativeSpecialDeduction(taxResult.getCumulativeSpecialDeduction());
        paySlip.setCumulativeStandardDeduction(taxResult.getCumulativeStandardDeduction());
        paySlip.setCumulativeTaxableIncome(taxResult.getCumulativeTaxableIncome());
        paySlip.setApplicableTaxRate(taxResult.getApplicableTaxRate());
        paySlip.setQuickDeduction(taxResult.getQuickDeduction());
        paySlip.setCumulativeTax(taxResult.getCumulativeTax());
        paySlip.setCumulativeTaxWithheld(taxResult.getCumulativeTaxWithheld());
        paySlip.setCurrentMonthTax(taxResult.getCurrentMonthTax());

        paySlip.setYearEndBonus(yearEndBonus);
        paySlip.setYearEndBonusTaxSeparately(yearEndBonusTaxSeparately);
        paySlip.setYearEndBonusTaxCombined(yearEndBonusTaxCombined);

        paySlip.setNetSalary(netSalary);

        dataStore.addPaySlip(paySlip);

        return paySlip;
    }

    private BigDecimal calculateProratedSalary(Employee employee, SalaryInput input) {
        BigDecimal baseSalary = employee.getEffectiveSalary(input.getYear(), input.getMonth());

        if (employee.isTerminatedThisMonth(input.getYear(), input.getMonth())
                && input.getActualWorkDays() != null
                && input.getTotalWorkDaysInMonth() != null
                && input.getTotalWorkDaysInMonth() > 0) {
            
            BigDecimal dailySalary = baseSalary.divide(
                    new BigDecimal(input.getTotalWorkDaysInMonth()), 10, RoundingMode.HALF_UP);
            return dailySalary.multiply(new BigDecimal(input.getActualWorkDays()))
                    .setScale(2, RoundingMode.HALF_UP);
        }

        return baseSalary;
    }

    private BigDecimal getAdditionalIncome(SalaryInput input, IncomeType type) {
        return input.getAdditionalIncomes().stream()
                .filter(income -> income.getType() == type)
                .map(AdditionalIncome::getAmount)
                .reduce(BigDecimal.ZERO, BigDecimal::add);
    }
}
