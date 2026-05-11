package com.company.payroll.service;

import com.company.payroll.model.PaySlip;

import java.io.BufferedWriter;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.List;

public class ExportService {

    public void exportPaySlipToText(PaySlip paySlip, String filePath) throws IOException {
        String content = paySlip.formatPaySlip();
        Path path = Paths.get(filePath);
        Files.write(path, content.getBytes(StandardCharsets.UTF_8));
    }

    public void exportPaySlipToCsv(PaySlip paySlip, String filePath) throws IOException {
        Path path = Paths.get(filePath);
        try (BufferedWriter writer = Files.newBufferedWriter(path, StandardCharsets.UTF_8)) {
            writer.write("字段,金额/内容\n");
            writer.write("员工," + paySlip.getEmployeeName() + "\n");
            writer.write("月份," + paySlip.getYear() + "年" + paySlip.getMonth() + "月\n");
            writer.write("基本工资," + paySlip.getBaseSalary() + "\n");
            writer.write("绩效奖金," + paySlip.getPerformanceBonus() + "\n");
            writer.write("加班费," + paySlip.getOvertimePay() + "\n");
            writer.write("项目奖金," + paySlip.getProjectBonus() + "\n");
            writer.write("兼职收入," + paySlip.getPartTimeIncome() + "\n");
            writer.write("交通补贴," + paySlip.getTransportationAllowance() + "\n");
            writer.write("通讯补贴," + paySlip.getCommunicationAllowance() + "\n");
            writer.write("餐补," + paySlip.getMealAllowance() + "\n");
            writer.write("税前总额," + paySlip.getGrossSalary() + "\n");
            writer.write("养老保险," + paySlip.getPensionInsurance() + "\n");
            writer.write("医疗保险," + paySlip.getMedicalInsurance() + "\n");
            writer.write("失业保险," + paySlip.getUnemploymentInsurance() + "\n");
            writer.write("住房公积金," + paySlip.getHousingFund() + "\n");
            writer.write("五险一金合计," + paySlip.getTotalSocialSecurityAndHousingFund() + "\n");
            writer.write("专项附加扣除," + paySlip.getSpecialDeductionTotal() + "\n");
            writer.write("累计应纳税所得额," + paySlip.getCumulativeTaxableIncome() + "\n");
            writer.write("适用税率," + paySlip.getApplicableTaxRate().multiply(new java.math.BigDecimal("100")) + "%\n");
            writer.write("速算扣除数," + paySlip.getQuickDeduction() + "\n");
            writer.write("本月应扣个税," + paySlip.getCurrentMonthTax() + "\n");
            if (paySlip.getYearEndBonus() != null && paySlip.getYearEndBonus().compareTo(java.math.BigDecimal.ZERO) > 0) {
                writer.write("年终奖," + paySlip.getYearEndBonus() + "\n");
                writer.write("年终奖单独计税," + paySlip.getYearEndBonusTaxSeparately() + "\n");
                writer.write("年终奖合并计税," + paySlip.getYearEndBonusTaxCombined() + "\n");
            }
            writer.write("税后实发," + paySlip.getNetSalary() + "\n");
        }
    }

    public void exportMultiplePaySlipsToCsv(List<PaySlip> paySlips, String filePath) throws IOException {
        Path path = Paths.get(filePath);
        try (BufferedWriter writer = Files.newBufferedWriter(path, StandardCharsets.UTF_8)) {
            writer.write("员工ID,员工姓名,年份,月份,税前总额,五险一金合计,专项附加扣除,本月个税,税后实发\n");
            for (PaySlip paySlip : paySlips) {
                writer.write(paySlip.getEmployeeId() + ",");
                writer.write(paySlip.getEmployeeName() + ",");
                writer.write(paySlip.getYear() + ",");
                writer.write(paySlip.getMonth() + ",");
                writer.write(paySlip.getGrossSalary() + ",");
                writer.write(paySlip.getTotalSocialSecurityAndHousingFund() + ",");
                writer.write(paySlip.getSpecialDeductionTotal() + ",");
                writer.write(paySlip.getCurrentMonthTax() + ",");
                writer.write(paySlip.getNetSalary() + "\n");
            }
        }
    }
}
