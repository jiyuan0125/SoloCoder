package com.payroll.client.util;

import com.payroll.common.dto.*;
import com.payroll.common.response.ApiResponse;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.Map;

public class OutputFormatter {

    private static final DateTimeFormatter DATE_FORMAT = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    public static void printArchive(PayrollArchiveDTO archive) {
        System.out.println("==========================================");
        System.out.println("归档ID: " + archive.getArchiveId());
        System.out.println("员工ID: " + archive.getEmployeeId());
        System.out.println("员工姓名: " + archive.getEmployeeName());
        System.out.println("身份证号: " + archive.getIdCard());
        System.out.println("年月: " + archive.getYear() + "-" + String.format("%02d", archive.getMonth()));
        System.out.println("基本工资: " + formatMoney(archive.getBaseSalary()));
        System.out.println("扣款明细:");
        if (archive.getDeductionDetails() != null && !archive.getDeductionDetails().isEmpty()) {
            for (DeductionDetailDTO detail : archive.getDeductionDetails()) {
                System.out.println("  - " + detail.getDeductionType() + ": " 
                    + formatMoney(detail.getAmount()) 
                    + (detail.getDescription() != null ? " (" + detail.getDescription() + ")" : ""));
            }
        }
        System.out.println("扣款总额: " + formatMoney(archive.getTotalDeduction()));
        System.out.println("实发金额: " + formatMoney(archive.getNetPay()));
        System.out.println("校验码: " + archive.getChecksum());
        System.out.println("数据有效性: " + (archive.isDataValid() ? "正常" : "异常"));
        System.out.println("状态: " + archive.getStatus());
        System.out.println("创建时间: " + (archive.getCreatedAt() != null ? archive.getCreatedAt().format(DATE_FORMAT) : "N/A"));
        System.out.println("确认时间: " + (archive.getConfirmedAt() != null ? archive.getConfirmedAt().format(DATE_FORMAT) : "N/A"));
        if (archive.getRemarks() != null && !archive.getRemarks().isEmpty()) {
            System.out.println("备注:");
            for (int i = 0; i < archive.getRemarks().size(); i++) {
                System.out.println("  [" + (i + 1) + "] " + archive.getRemarks().get(i));
            }
        }
        System.out.println("==========================================");
    }

    public static void printArchiveList(List<PayrollArchiveDTO> archives) {
        if (archives == null || archives.isEmpty()) {
            System.out.println("未找到任何归档记录");
            return;
        }
        System.out.println("==========================================");
        System.out.println("共找到 " + archives.size() + " 条归档记录");
        System.out.println("------------------------------------------");
        for (PayrollArchiveDTO archive : archives) {
            System.out.println("归档ID: " + archive.getArchiveId());
            System.out.println("员工: " + archive.getEmployeeId() + " - " + archive.getEmployeeName());
            System.out.println("年月: " + archive.getYear() + "-" + String.format("%02d", archive.getMonth()));
            System.out.println("实发金额: " + formatMoney(archive.getNetPay()));
            System.out.println("数据有效性: " + (archive.isDataValid() ? "正常" : "异常"));
            System.out.println("状态: " + archive.getStatus());
            System.out.println("校验码: " + archive.getChecksum());
            System.out.println("------------------------------------------");
        }
    }

    public static void printAnnualReport(AnnualReportDTO report) {
        System.out.println("==========================================");
        System.out.println("年度薪资汇总报表 - " + report.getYear() + "年");
        System.out.println("==========================================");
        System.out.println("总支出: " + formatMoney(report.getTotalExpense()));
        System.out.println("人均薪资: " + formatMoney(report.getAverageSalary()));
        System.out.println("最高薪资: " + formatMoney(report.getMaxSalary()));
        System.out.println("最低薪资: " + formatMoney(report.getMinSalary()));
        System.out.println("总记录数: " + report.getTotalRecords());
        System.out.println("员工人数: " + report.getTotalEmployees());
        System.out.println("同比变化率: " + report.getYearOverYearChange() + "%");
        
        if (report.getMonthlySummaries() != null && !report.getMonthlySummaries().isEmpty()) {
            System.out.println("\n月度明细:");
            System.out.println("------------------------------------------");
            System.out.printf("%-6s | %-15s | %-10s\n", "月份", "总支出", "记录数");
            System.out.println("------------------------------------------");
            for (MonthlySummaryDTO summary : report.getMonthlySummaries()) {
                System.out.printf("%-6d | %-15s | %-10d\n",
                        summary.getMonth(),
                        formatMoney(summary.getTotalExpense()),
                        summary.getRecordCount());
            }
        }
        System.out.println("==========================================");
    }

    public static void printYearComparison(YearComparisonDTO comparison) {
        System.out.println("==========================================");
        System.out.println("跨年对比分析");
        System.out.println("==========================================");
        System.out.println("基准年: " + comparison.getBaseYear());
        System.out.println("对比年: " + comparison.getCompareYear());
        System.out.println("------------------------------------------");
        System.out.println("基准年总支出: " + formatMoney(comparison.getBaseYearTotal()));
        System.out.println("对比年总支出: " + formatMoney(comparison.getCompareYearTotal()));
        System.out.println("变化金额: " + formatMoney(comparison.getTotalChange()));
        System.out.println("变化率: " + comparison.getTotalChangeRate() + "%");
        
        if (comparison.getMonthlyComparisons() != null && !comparison.getMonthlyComparisons().isEmpty()) {
            System.out.println("\n月度对比:");
            System.out.println("------------------------------------------");
            System.out.printf("%-4s | %-12s | %-12s | %-8s | %-8s\n",
                    "月份", "基准年", "对比年", "变化", "变化率");
            System.out.println("------------------------------------------");
            for (MonthlyComparisonDTO mc : comparison.getMonthlyComparisons()) {
                System.out.printf("%-4d | %-12s | %-12s | %-8s | %-8s%%\n",
                        mc.getMonth(),
                        formatMoney(mc.getBaseYearAmount()),
                        formatMoney(mc.getCompareYearAmount()),
                        formatMoney(mc.getChange()),
                        mc.getChangeRate().setScale(2, RoundingMode.HALF_UP).toPlainString());
            }
        }
        System.out.println("==========================================");
    }

    public static void printSuccess(String message) {
        System.out.println("[成功] " + message);
    }

    public static void printError(int code, String message) {
        System.err.println("[错误] 代码: " + code + ", 信息: " + message);
    }

    public static void printHelp() {
        System.out.println("==========================================");
        System.out.println("薪资归档系统 - 命令行客户端");
        System.out.println("==========================================");
        System.out.println("\n可用命令:");
        System.out.println("  create <json>          创建归档记录");
        System.out.println("  confirm <archiveId>    确认归档");
        System.out.println("  remark <archiveId> <text> 追加备注");
        System.out.println("  get <archiveId>        查询单条归档");
        System.out.println("  query [options]        查询归档列表");
        System.out.println("    选项:");
        System.out.println("      --employee=<id>    按员工ID筛选");
        System.out.println("      --year=<year>      按年份筛选");
        System.out.println("      --month=<month>    按月份筛选");
        System.out.println("  report <year>          生成年度报表");
        System.out.println("  compare <base> <comp>  跨年对比分析");
        System.out.println("  help                   显示帮助信息");
        System.out.println("\n示例:");
        System.out.println("  payroll-client query --year=2024 --month=1");
        System.out.println("  payroll-client report 2024");
        System.out.println("  payroll-client compare 2023 2024");
        System.out.println("==========================================");
    }

    private static String formatMoney(BigDecimal amount) {
        if (amount == null) {
            return "¥0.00";
        }
        return "¥" + amount.setScale(2, RoundingMode.HALF_UP).toPlainString();
    }
}
