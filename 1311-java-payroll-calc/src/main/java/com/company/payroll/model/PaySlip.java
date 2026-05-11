package com.company.payroll.model;

import java.math.BigDecimal;
import java.math.RoundingMode;

public class PaySlip {
    private String id;
    private String employeeId;
    private String employeeName;
    private int year;
    private int month;
    
    private BigDecimal grossSalary;
    private BigDecimal baseSalary;
    private BigDecimal performanceBonus;
    private BigDecimal overtimePay;
    private BigDecimal projectBonus;
    private BigDecimal partTimeIncome;
    private BigDecimal transportationAllowance;
    private BigDecimal communicationAllowance;
    private BigDecimal mealAllowance;
    
    private BigDecimal taxableAllowance;
    
    private BigDecimal pensionInsurance;
    private BigDecimal medicalInsurance;
    private BigDecimal unemploymentInsurance;
    private BigDecimal housingFund;
    private BigDecimal totalSocialSecurityAndHousingFund;
    
    private BigDecimal specialDeductionTotal;
    
    private BigDecimal cumulativeIncome;
    private BigDecimal cumulativeSocialSecurity;
    private BigDecimal cumulativeSpecialDeduction;
    private BigDecimal cumulativeStandardDeduction;
    private BigDecimal cumulativeTaxableIncome;
    
    private BigDecimal applicableTaxRate;
    private BigDecimal quickDeduction;
    private BigDecimal cumulativeTax;
    private BigDecimal cumulativeTaxWithheld;
    private BigDecimal currentMonthTax;
    
    private BigDecimal yearEndBonus;
    private BigDecimal yearEndBonusTaxSeparately;
    private BigDecimal yearEndBonusTaxCombined;
    
    private BigDecimal netSalary;

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getEmployeeId() {
        return employeeId;
    }

    public void setEmployeeId(String employeeId) {
        this.employeeId = employeeId;
    }

    public String getEmployeeName() {
        return employeeName;
    }

    public void setEmployeeName(String employeeName) {
        this.employeeName = employeeName;
    }

    public int getYear() {
        return year;
    }

    public void setYear(int year) {
        this.year = year;
    }

    public int getMonth() {
        return month;
    }

    public void setMonth(int month) {
        this.month = month;
    }

    public BigDecimal getGrossSalary() {
        return grossSalary;
    }

    public void setGrossSalary(BigDecimal grossSalary) {
        this.grossSalary = grossSalary;
    }

    public BigDecimal getBaseSalary() {
        return baseSalary;
    }

    public void setBaseSalary(BigDecimal baseSalary) {
        this.baseSalary = baseSalary;
    }

    public BigDecimal getPerformanceBonus() {
        return performanceBonus;
    }

    public void setPerformanceBonus(BigDecimal performanceBonus) {
        this.performanceBonus = performanceBonus;
    }

    public BigDecimal getOvertimePay() {
        return overtimePay;
    }

    public void setOvertimePay(BigDecimal overtimePay) {
        this.overtimePay = overtimePay;
    }

    public BigDecimal getProjectBonus() {
        return projectBonus;
    }

    public void setProjectBonus(BigDecimal projectBonus) {
        this.projectBonus = projectBonus;
    }

    public BigDecimal getPartTimeIncome() {
        return partTimeIncome;
    }

    public void setPartTimeIncome(BigDecimal partTimeIncome) {
        this.partTimeIncome = partTimeIncome;
    }

    public BigDecimal getTransportationAllowance() {
        return transportationAllowance;
    }

    public void setTransportationAllowance(BigDecimal transportationAllowance) {
        this.transportationAllowance = transportationAllowance;
    }

    public BigDecimal getCommunicationAllowance() {
        return communicationAllowance;
    }

    public void setCommunicationAllowance(BigDecimal communicationAllowance) {
        this.communicationAllowance = communicationAllowance;
    }

    public BigDecimal getMealAllowance() {
        return mealAllowance;
    }

    public void setMealAllowance(BigDecimal mealAllowance) {
        this.mealAllowance = mealAllowance;
    }

    public BigDecimal getTaxableAllowance() {
        return taxableAllowance;
    }

    public void setTaxableAllowance(BigDecimal taxableAllowance) {
        this.taxableAllowance = taxableAllowance;
    }

    public BigDecimal getPensionInsurance() {
        return pensionInsurance;
    }

    public void setPensionInsurance(BigDecimal pensionInsurance) {
        this.pensionInsurance = pensionInsurance;
    }

    public BigDecimal getMedicalInsurance() {
        return medicalInsurance;
    }

    public void setMedicalInsurance(BigDecimal medicalInsurance) {
        this.medicalInsurance = medicalInsurance;
    }

    public BigDecimal getUnemploymentInsurance() {
        return unemploymentInsurance;
    }

    public void setUnemploymentInsurance(BigDecimal unemploymentInsurance) {
        this.unemploymentInsurance = unemploymentInsurance;
    }

    public BigDecimal getHousingFund() {
        return housingFund;
    }

    public void setHousingFund(BigDecimal housingFund) {
        this.housingFund = housingFund;
    }

    public BigDecimal getTotalSocialSecurityAndHousingFund() {
        return totalSocialSecurityAndHousingFund;
    }

    public void setTotalSocialSecurityAndHousingFund(BigDecimal totalSocialSecurityAndHousingFund) {
        this.totalSocialSecurityAndHousingFund = totalSocialSecurityAndHousingFund;
    }

    public BigDecimal getSpecialDeductionTotal() {
        return specialDeductionTotal;
    }

    public void setSpecialDeductionTotal(BigDecimal specialDeductionTotal) {
        this.specialDeductionTotal = specialDeductionTotal;
    }

    public BigDecimal getCumulativeIncome() {
        return cumulativeIncome;
    }

    public void setCumulativeIncome(BigDecimal cumulativeIncome) {
        this.cumulativeIncome = cumulativeIncome;
    }

    public BigDecimal getCumulativeSocialSecurity() {
        return cumulativeSocialSecurity;
    }

    public void setCumulativeSocialSecurity(BigDecimal cumulativeSocialSecurity) {
        this.cumulativeSocialSecurity = cumulativeSocialSecurity;
    }

    public BigDecimal getCumulativeSpecialDeduction() {
        return cumulativeSpecialDeduction;
    }

    public void setCumulativeSpecialDeduction(BigDecimal cumulativeSpecialDeduction) {
        this.cumulativeSpecialDeduction = cumulativeSpecialDeduction;
    }

    public BigDecimal getCumulativeStandardDeduction() {
        return cumulativeStandardDeduction;
    }

    public void setCumulativeStandardDeduction(BigDecimal cumulativeStandardDeduction) {
        this.cumulativeStandardDeduction = cumulativeStandardDeduction;
    }

    public BigDecimal getCumulativeTaxableIncome() {
        return cumulativeTaxableIncome;
    }

    public void setCumulativeTaxableIncome(BigDecimal cumulativeTaxableIncome) {
        this.cumulativeTaxableIncome = cumulativeTaxableIncome;
    }

    public BigDecimal getApplicableTaxRate() {
        return applicableTaxRate;
    }

    public void setApplicableTaxRate(BigDecimal applicableTaxRate) {
        this.applicableTaxRate = applicableTaxRate;
    }

    public BigDecimal getQuickDeduction() {
        return quickDeduction;
    }

    public void setQuickDeduction(BigDecimal quickDeduction) {
        this.quickDeduction = quickDeduction;
    }

    public BigDecimal getCumulativeTax() {
        return cumulativeTax;
    }

    public void setCumulativeTax(BigDecimal cumulativeTax) {
        this.cumulativeTax = cumulativeTax;
    }

    public BigDecimal getCumulativeTaxWithheld() {
        return cumulativeTaxWithheld;
    }

    public void setCumulativeTaxWithheld(BigDecimal cumulativeTaxWithheld) {
        this.cumulativeTaxWithheld = cumulativeTaxWithheld;
    }

    public BigDecimal getCurrentMonthTax() {
        return currentMonthTax;
    }

    public void setCurrentMonthTax(BigDecimal currentMonthTax) {
        this.currentMonthTax = currentMonthTax;
    }

    public BigDecimal getYearEndBonus() {
        return yearEndBonus;
    }

    public void setYearEndBonus(BigDecimal yearEndBonus) {
        this.yearEndBonus = yearEndBonus;
    }

    public BigDecimal getYearEndBonusTaxSeparately() {
        return yearEndBonusTaxSeparately;
    }

    public void setYearEndBonusTaxSeparately(BigDecimal yearEndBonusTaxSeparately) {
        this.yearEndBonusTaxSeparately = yearEndBonusTaxSeparately;
    }

    public BigDecimal getYearEndBonusTaxCombined() {
        return yearEndBonusTaxCombined;
    }

    public void setYearEndBonusTaxCombined(BigDecimal yearEndBonusTaxCombined) {
        this.yearEndBonusTaxCombined = yearEndBonusTaxCombined;
    }

    public BigDecimal getNetSalary() {
        return netSalary;
    }

    public void setNetSalary(BigDecimal netSalary) {
        this.netSalary = netSalary;
    }

    public String formatPaySlip() {
        StringBuilder sb = new StringBuilder();
        sb.append("=========================== 工资条 ===========================\n");
        sb.append("员工: ").append(employeeName).append(" (").append(employeeId).append(")\n");
        sb.append("月份: ").append(year).append("年").append(month).append("月\n");
        sb.append("-------------------------------------------------------------\n");
        sb.append("【收入明细】\n");
        sb.append("  基本工资: ").append(formatMoney(baseSalary)).append("\n");
        if (performanceBonus != null && performanceBonus.compareTo(BigDecimal.ZERO) > 0) {
            sb.append("  绩效奖金: ").append(formatMoney(performanceBonus)).append("\n");
        }
        if (overtimePay != null && overtimePay.compareTo(BigDecimal.ZERO) > 0) {
            sb.append("  加班费: ").append(formatMoney(overtimePay)).append("\n");
        }
        if (projectBonus != null && projectBonus.compareTo(BigDecimal.ZERO) > 0) {
            sb.append("  项目奖金: ").append(formatMoney(projectBonus)).append("\n");
        }
        if (partTimeIncome != null && partTimeIncome.compareTo(BigDecimal.ZERO) > 0) {
            sb.append("  兼职收入: ").append(formatMoney(partTimeIncome)).append("\n");
        }
        if (transportationAllowance != null && transportationAllowance.compareTo(BigDecimal.ZERO) > 0) {
            sb.append("  交通补贴: ").append(formatMoney(transportationAllowance)).append("\n");
        }
        if (communicationAllowance != null && communicationAllowance.compareTo(BigDecimal.ZERO) > 0) {
            sb.append("  通讯补贴: ").append(formatMoney(communicationAllowance)).append("\n");
        }
        if (mealAllowance != null && mealAllowance.compareTo(BigDecimal.ZERO) > 0) {
            sb.append("  餐补: ").append(formatMoney(mealAllowance)).append("\n");
        }
        if (taxableAllowance != null && taxableAllowance.compareTo(BigDecimal.ZERO) > 0) {
            sb.append("  补贴应税部分: ").append(formatMoney(taxableAllowance)).append("\n");
        }
        if (yearEndBonus != null && yearEndBonus.compareTo(BigDecimal.ZERO) > 0) {
            sb.append("  年终奖: ").append(formatMoney(yearEndBonus)).append("\n");
        }
        sb.append("-------------------------------------------------------------\n");
        sb.append("【税前总额】: ").append(formatMoney(grossSalary)).append("\n");
        sb.append("-------------------------------------------------------------\n");
        sb.append("【五险一金个人部分】\n");
        sb.append("  养老保险: ").append(formatMoney(pensionInsurance)).append("\n");
        sb.append("  医疗保险: ").append(formatMoney(medicalInsurance)).append("\n");
        sb.append("  失业保险: ").append(formatMoney(unemploymentInsurance)).append("\n");
        sb.append("  住房公积金: ").append(formatMoney(housingFund)).append("\n");
        sb.append("  合计: ").append(formatMoney(totalSocialSecurityAndHousingFund)).append("\n");
        sb.append("-------------------------------------------------------------\n");
        sb.append("【专项附加扣除总额】: ").append(formatMoney(specialDeductionTotal)).append("\n");
        sb.append("-------------------------------------------------------------\n");
        sb.append("【个税计算】\n");
        sb.append("  累计收入: ").append(formatMoney(cumulativeIncome)).append("\n");
        sb.append("  累计五险一金: ").append(formatMoney(cumulativeSocialSecurity)).append("\n");
        sb.append("  累计专项附加扣除: ").append(formatMoney(cumulativeSpecialDeduction)).append("\n");
        sb.append("  累计减除费用: ").append(formatMoney(cumulativeStandardDeduction)).append("\n");
        sb.append("  累计应纳税所得额: ").append(formatMoney(cumulativeTaxableIncome)).append("\n");
        sb.append("  适用税率: ").append(applicableTaxRate.multiply(new BigDecimal("100")).setScale(0)).append("%\n");
        sb.append("  速算扣除数: ").append(formatMoney(quickDeduction)).append("\n");
        sb.append("  累计应纳税额: ").append(formatMoney(cumulativeTax)).append("\n");
        sb.append("  累计已预扣税额: ").append(formatMoney(cumulativeTaxWithheld)).append("\n");
        sb.append("  本月应扣个税: ").append(formatMoney(currentMonthTax)).append("\n");
        sb.append("-------------------------------------------------------------\n");
        if (yearEndBonus != null && yearEndBonus.compareTo(BigDecimal.ZERO) > 0) {
            sb.append("【年终奖计税方案】\n");
            sb.append("  方案一（单独计税）税额: ").append(formatMoney(yearEndBonusTaxSeparately)).append("\n");
            sb.append("  方案二（合并计税）税额: ").append(formatMoney(yearEndBonusTaxCombined)).append("\n");
            sb.append("  建议选择: ");
            if (yearEndBonusTaxSeparately.compareTo(yearEndBonusTaxCombined) <= 0) {
                sb.append("方案一（单独计税更划算）\n");
            } else {
                sb.append("方案二（合并计税更划算）\n");
            }
            sb.append("-------------------------------------------------------------\n");
        }
        sb.append("【税后实发】: ").append(formatMoney(netSalary)).append("\n");
        sb.append("=============================================================\n");
        return sb.toString();
    }

    private String formatMoney(BigDecimal amount) {
        if (amount == null) {
            return "¥0.00";
        }
        return "¥" + amount.setScale(2, RoundingMode.HALF_UP).toString();
    }
}
