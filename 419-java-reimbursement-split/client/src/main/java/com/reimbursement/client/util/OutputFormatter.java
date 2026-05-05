package com.reimbursement.client.util;

import com.reimbursement.common.dto.*;

import java.math.BigDecimal;
import java.time.format.DateTimeFormatter;
import java.util.List;

public class OutputFormatter {
    private static final DateTimeFormatter DATE_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
    private static final String LINE = "----------------------------------------------------------------";

    public static void printApiResponse(ApiResponse<?> response) {
        System.out.println();
        System.out.println(LINE);
        if (response.getCode() == 200) {
            System.out.println("\u2714 操作成功: " + response.getMessage());
        } else {
            System.out.println("\u274c 操作失败 (错误码: " + response.getCode() + ")");
            System.out.println("   错误信息: " + response.getMessage());
        }
        System.out.println(LINE);
    }

    public static void printReimbursementDetail(ReimbursementDetailDTO dto) {
        System.out.println();
        System.out.println(LINE);
        System.out.println("                    报销单详情");
        System.out.println(LINE);
        System.out.printf("报销单号: %s%n", dto.getReimbursementId());
        System.out.printf("员工ID: %s  姓名: %s%n", dto.getEmployeeId(), dto.getEmployeeName());
        System.out.printf("报销类型: %s%n", dto.getReimbursementType());
        System.out.printf("描述: %s%n", dto.getDescription());
        System.out.printf("总金额: ¥%s%n", formatAmount(dto.getTotalAmount()));
        System.out.printf("状态: %s%n", dto.getStatus().getDescription());
        System.out.printf("创建时间: %s%n", dto.getCreateTime().format(DATE_FORMATTER));
        System.out.printf("更新时间: %s%n", dto.getUpdateTime().format(DATE_FORMATTER));
        
        if (dto.getRejectReason() != null && !dto.getRejectReason().isEmpty()) {
            System.out.printf("驳回原因: %s%n", dto.getRejectReason());
        }
        
        System.out.println();
        System.out.println("【成本中心分配明细】");
        System.out.printf("%-12s %-12s %-8s %-12s %-10s%n", 
            "成本中心ID", "成本中心", "分配比例", "分配金额", "审批状态");
        System.out.println("------------------------------------------------------------");
        
        for (CostCenterApprovalDTO allocation : dto.getAllocations()) {
            String status = allocation.getApprovalStatus().getDescription();
            if (allocation.getApprovalStatus() == com.reimbursement.common.enums.ApprovalStatus.APPROVED) {
                status = "\u2714 " + status;
            } else if (allocation.getApprovalStatus() == com.reimbursement.common.enums.ApprovalStatus.REJECTED) {
                status = "\u274c " + status;
            }
            
            System.out.printf("%-12s %-12s %-10d %-14s %s%n",
                allocation.getCostCenterId(),
                allocation.getCostCenterName(),
                allocation.getPercentage(),
                "¥" + formatAmount(allocation.getAllocatedAmount()),
                status);
            
            if (allocation.getApproverName() != null) {
                System.out.printf("  审批人: %s (%s)%n", allocation.getApproverName(), 
                    allocation.getApprovalTime() != null ? allocation.getApprovalTime().format(DATE_FORMATTER) : "-");
            }
            if (allocation.getComment() != null && !allocation.getComment().isEmpty()) {
                System.out.printf("  备注: %s%n", allocation.getComment());
            }
        }
        System.out.println(LINE);
    }

    public static void printReimbursementList(List<ReimbursementDetailDTO> list) {
        System.out.println();
        System.out.println(LINE);
        System.out.println("                    报销单列表");
        System.out.println(LINE);
        System.out.printf("%-14s %-10s %-10s %-14s %-12s%n",
            "报销单号", "员工", "类型", "金额", "状态");
        System.out.println("------------------------------------------------------------");
        
        for (ReimbursementDetailDTO dto : list) {
            System.out.printf("%-14s %-10s %-10s %-14s %s%n",
                dto.getReimbursementId(),
                dto.getEmployeeName(),
                dto.getReimbursementType(),
                "¥" + formatAmount(dto.getTotalAmount()),
                dto.getStatus().getDescription());
        }
        System.out.println(LINE);
        System.out.printf("共 %d 条记录%n", list.size());
    }

    public static void printCostCenterList(List<CostCenterDTO> list) {
        System.out.println();
        System.out.println(LINE);
        System.out.println("                    成本中心列表");
        System.out.println(LINE);
        System.out.printf("%-12s %-12s %-10s %-15s%n",
            "成本中心ID", "名称", "负责人", "月度预算");
        System.out.println("------------------------------------------------------------");
        
        for (CostCenterDTO cc : list) {
            System.out.printf("%-12s %-12s %-10s ¥%-14s%n",
                cc.getId(),
                cc.getName(),
                cc.getManagerName(),
                formatAmount(cc.getMonthlyBudget()));
        }
        System.out.println(LINE);
    }

    public static void printMonthlyReport(MonthlyReportDTO report) {
        System.out.println();
        System.out.println(LINE);
        System.out.printf("                    %d年%d月月度支出报表%n", report.getYear(), report.getMonth());
        System.out.println(LINE);
        System.out.printf("成本中心: %s (%s)%n", report.getCostCenterName(), report.getCostCenterId());
        System.out.printf("月度预算: ¥%s%n", formatAmount(report.getBudgetAmount()));
        System.out.printf("已支出:   ¥%s%n", formatAmount(report.getTotalAmount()));
        System.out.printf("剩余预算: ¥%s%n", formatAmount(report.getRemainingBudget()));
        
        if (report.isOverBudget()) {
            System.out.println("\u26a0  【警告】已超支!");
        } else if (report.getRemainingBudget().compareTo(report.getBudgetAmount().multiply(new BigDecimal("0.1"))) < 0) {
            System.out.println("\u26a0  【提醒】预算不足10%");
        }
        
        if (report.getReimbursements() != null && !report.getReimbursements().isEmpty()) {
            System.out.println();
            System.out.println("【报销明细】");
            System.out.printf("%-14s %-10s %-10s %-14s%n",
                "报销单号", "员工", "类型", "金额");
            System.out.println("------------------------------------------------------------");
            
            for (MonthlyReportDTO.ReimbursementSummaryDTO item : report.getReimbursements()) {
                System.out.printf("%-14s %-10s %-10s ¥%-14s%n",
                    item.getReimbursementId(),
                    item.getEmployeeName(),
                    item.getReimbursementType(),
                    formatAmount(item.getAmount()));
            }
        }
        System.out.println(LINE);
    }

    public static void printHelp() {
        System.out.println();
        System.out.println(LINE);
        System.out.println("                    报销拆分系统 - 命令帮助");
        System.out.println(LINE);
        System.out.println();
        System.out.println("【查询命令】");
        System.out.println("  list                     - 列出所有报销单");
        System.out.println("  get <报销单号>           - 查询报销单详情");
        System.out.println("  cost-centers             - 列出所有成本中心");
        System.out.println("  report <年> <月> <成本中心ID> - 查询月度支出报表");
        System.out.println();
        System.out.println("【操作命令】");
        System.out.println("  create                   - 创建报销单（交互式）");
        System.out.println("  submit <报销单号>        - 提交报销单");
        System.out.println("  approve                  - 审批报销单（交互式）");
        System.out.println("  final-review             - 财务复核（交互式）");
        System.out.println();
        System.out.println("【示例】");
        System.out.println("  list");
        System.out.println("  get RB1234567890AB");
        System.out.println("  cost-centers");
        System.out.println("  report 2026 5 CC001");
        System.out.println("  create");
        System.out.println(LINE);
    }

    private static String formatAmount(BigDecimal amount) {
        if (amount == null) {
            return "0.00";
        }
        return amount.setScale(2, java.math.RoundingMode.HALF_UP).toString();
    }
}
