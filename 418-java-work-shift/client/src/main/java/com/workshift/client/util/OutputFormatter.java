package com.workshift.client.util;

import com.workshift.common.dto.EmployeeDTO;
import com.workshift.common.dto.MonthlyStatisticsDTO;
import com.workshift.common.dto.ObjectionDTO;
import com.workshift.common.dto.ShiftDTO;
import com.workshift.common.dto.ShiftSwapDTO;
import com.workshift.common.enums.ShiftType;
import java.util.List;
import java.util.Map;

public class OutputFormatter {

    public static void printSeparator() {
        System.out.println("----------------------------------------------------------------");
    }

    public static void printHeader(String title) {
        System.out.println("╔══════════════════════════════════════════════════════════════╗");
        System.out.println("║  " + padRight(title, 58) + "║");
        System.out.println("╚══════════════════════════════════════════════════════════════╝");
    }

    public static void printSuccess(String message) {
        System.out.println("✅ " + message);
    }

    public static void printError(String message) {
        System.out.println("❌ 错误: " + message);
    }

    public static void printEmployees(List<EmployeeDTO> employees) {
        if (employees.isEmpty()) {
            System.out.println("没有找到员工记录");
            return;
        }
        
        System.out.println("员工列表:");
        printSeparator();
        System.out.printf("%-36s %-15s %-15s %-10s%n", "ID", "姓名", "部门", "时薪");
        printSeparator();
        
        for (EmployeeDTO emp : employees) {
            System.out.printf("%-36s %-15s %-15s %-10s%n",
                    emp.getId(),
                    emp.getName(),
                    emp.getDepartment(),
                    emp.getHourlyWage() != null ? emp.getHourlyWage().toString() : "N/A");
        }
        printSeparator();
        System.out.println("总计: " + employees.size() + " 名员工");
    }

    public static void printEmployee(EmployeeDTO emp) {
        if (emp == null) {
            System.out.println("没有找到员工记录");
            return;
        }
        
        System.out.println("员工详情:");
        printSeparator();
        System.out.println("  ID:         " + emp.getId());
        System.out.println("  姓名:       " + emp.getName());
        System.out.println("  部门:       " + emp.getDepartment());
        System.out.println("  时薪:       " + emp.getHourlyWage());
        printSeparator();
    }

    public static void printShifts(List<ShiftDTO> shifts) {
        if (shifts.isEmpty()) {
            System.out.println("没有找到排班记录");
            return;
        }
        
        System.out.println("排班列表:");
        printSeparator();
        System.out.printf("%-36s %-15s %-12s %-10s %-8s %-8s%n", 
                "ID", "员工ID", "日期", "班次", "已发布", "节假日");
        printSeparator();
        
        for (ShiftDTO shift : shifts) {
            System.out.printf("%-36s %-15s %-12s %-10s %-8s %-8s%n",
                    shift.getId(),
                    shift.getEmployeeId(),
                    shift.getDate(),
                    formatShiftType(shift.getShiftType()),
                    shift.isPublished() ? "是" : "否",
                    shift.isHoliday() ? "是" : "否");
        }
        printSeparator();
        System.out.println("总计: " + shifts.size() + " 条排班记录");
    }

    public static void printShift(ShiftDTO shift) {
        if (shift == null) {
            System.out.println("没有找到排班记录");
            return;
        }
        
        System.out.println("排班详情:");
        printSeparator();
        System.out.println("  ID:           " + shift.getId());
        System.out.println("  员工ID:       " + shift.getEmployeeId());
        System.out.println("  日期:         " + shift.getDate());
        System.out.println("  班次:         " + formatShiftType(shift.getShiftType()));
        System.out.println("  工作时长:     " + shift.getWorkHours() + " 小时");
        System.out.println("  已发布:       " + (shift.isPublished() ? "是" : "否"));
        System.out.println("  发布时间:     " + (shift.getPublishTime() != null ? shift.getPublishTime() : "N/A"));
        System.out.println("  节假日:       " + (shift.isHoliday() ? "是" : "否"));
        printSeparator();
    }

    public static void printSwaps(List<ShiftSwapDTO> swaps) {
        if (swaps.isEmpty()) {
            System.out.println("没有找到换班申请记录");
            return;
        }
        
        System.out.println("换班申请列表:");
        printSeparator();
        System.out.printf("%-36s %-12s %-12s %-12s %-12s %-10s%n", 
                "ID", "申请人", "目标人", "申请日期", "目标日期", "状态");
        printSeparator();
        
        for (ShiftSwapDTO swap : swaps) {
            System.out.printf("%-36s %-12s %-12s %-12s %-12s %-10s%n",
                    swap.getId(),
                    swap.getRequesterEmployeeId(),
                    swap.getTargetEmployeeId(),
                    swap.getRequesterDate(),
                    swap.getTargetDate(),
                    swap.getStatus() != null ? swap.getStatus().getDescription() : "N/A");
        }
        printSeparator();
        System.out.println("总计: " + swaps.size() + " 条换班申请");
    }

    public static void printSwap(ShiftSwapDTO swap) {
        if (swap == null) {
            System.out.println("没有找到换班申请记录");
            return;
        }
        
        System.out.println("换班申请详情:");
        printSeparator();
        System.out.println("  ID:               " + swap.getId());
        System.out.println("  申请人ID:         " + swap.getRequesterEmployeeId());
        System.out.println("  申请日期:         " + swap.getRequesterDate());
        System.out.println("  申请班次:         " + formatShiftType(swap.getRequesterShiftType()));
        System.out.println("  目标员工ID:       " + swap.getTargetEmployeeId());
        System.out.println("  目标日期:         " + swap.getTargetDate());
        System.out.println("  目标班次:         " + formatShiftType(swap.getTargetShiftType()));
        System.out.println("  状态:             " + (swap.getStatus() != null ? swap.getStatus().getDescription() : "N/A"));
        System.out.println("  申请人签名:       " + (swap.getRequesterSignature() != null ? swap.getRequesterSignature() : "N/A"));
        System.out.println("  目标人签名:       " + (swap.getTargetSignature() != null ? swap.getTargetSignature() : "N/A"));
        System.out.println("  拒绝原因:         " + (swap.getRejectReason() != null ? swap.getRejectReason() : "N/A"));
        System.out.println("  创建时间:         " + swap.getCreateTime());
        printSeparator();
    }

    public static void printObjections(List<ObjectionDTO> objections) {
        if (objections.isEmpty()) {
            System.out.println("没有找到异议记录");
            return;
        }
        
        System.out.println("异议列表:");
        printSeparator();
        System.out.printf("%-36s %-36s %-12s %-10s%n", 
                "ID", "排班ID", "员工ID", "状态");
        printSeparator();
        
        for (ObjectionDTO obj : objections) {
            System.out.printf("%-36s %-36s %-12s %-10s%n",
                    obj.getId(),
                    obj.getShiftId(),
                    obj.getEmployeeId(),
                    obj.getStatus() != null ? obj.getStatus().getDescription() : "N/A");
        }
        printSeparator();
        System.out.println("总计: " + objections.size() + " 条异议记录");
    }

    public static void printObjection(ObjectionDTO obj) {
        if (obj == null) {
            System.out.println("没有找到异议记录");
            return;
        }
        
        System.out.println("异议详情:");
        printSeparator();
        System.out.println("  ID:               " + obj.getId());
        System.out.println("  排班ID:           " + obj.getShiftId());
        System.out.println("  员工ID:           " + obj.getEmployeeId());
        System.out.println("  异议原因:         " + obj.getReason());
        System.out.println("  状态:             " + (obj.getStatus() != null ? obj.getStatus().getDescription() : "N/A"));
        System.out.println("  处理备注:         " + (obj.getHandlerNote() != null ? obj.getHandlerNote() : "N/A"));
        System.out.println("  创建时间:         " + obj.getCreateTime());
        System.out.println("  处理时间:         " + (obj.getHandleTime() != null ? obj.getHandleTime() : "N/A"));
        printSeparator();
    }

    public static void printMonthlyStatistics(MonthlyStatisticsDTO stats) {
        if (stats == null) {
            System.out.println("没有找到统计数据");
            return;
        }
        
        System.out.println("月度排班统计报表:");
        printSeparator();
        System.out.println("  员工ID:           " + stats.getEmployeeId());
        System.out.println("  统计月份:         " + stats.getYearMonth());
        printSeparator();
        System.out.println("【工作统计】");
        System.out.println("  总工作天数:       " + stats.getTotalWorkingDays() + " 天");
        System.out.println("  总休息天数:       " + stats.getTotalRestDays() + " 天");
        System.out.println("  早班次数:         " + stats.getMorningShiftCount());
        System.out.println("  中班次数:         " + stats.getAfternoonShiftCount());
        System.out.println("  夜班次数:         " + stats.getNightShiftCount());
        System.out.println("  节假日工作天数:   " + stats.getHolidayWorkingDays() + " 天");
        printSeparator();
        System.out.println("【工时统计】");
        System.out.println("  正常工作时长:     " + stats.getNormalWorkHours() + " 小时");
        System.out.println("  节假日工作时长:   " + stats.getHolidayWorkHours() + " 小时");
        System.out.println("  总工作时长:       " + stats.getTotalWorkHours() + " 小时");
        printSeparator();
        System.out.println("【薪资统计】");
        System.out.println("  正常工资:         " + stats.getNormalWage());
        System.out.println("  节假日加班费:     " + stats.getHolidayOvertimeWage());
        System.out.println("  总工资:           " + stats.getTotalWage());
        printSeparator();
        System.out.println("【其他统计】");
        System.out.println("  换班次数:         " + stats.getSwapCount());
        System.out.println("  异议次数:         " + stats.getObjectionCount());
        System.out.println("  最大连续夜班:     " + stats.getConsecutiveNightShiftMax() + " 天");
        System.out.println("  最大连续工作:     " + stats.getConsecutiveWorkDaysMax() + " 天");
        printSeparator();
        System.out.println("【合规性检查】");
        System.out.println("  规则合规:         " + (stats.isRuleCompliant() ? "是" : "否"));
        if (stats.getViolations() != null && !stats.getViolations().isEmpty()) {
            System.out.println("  违规项:");
            for (Map.Entry<String, Object> entry : stats.getViolations().entrySet()) {
                System.out.println("    - " + entry.getKey() + ": " + entry.getValue());
            }
        }
        printSeparator();
    }

    public static void printAllMonthlyStatistics(List<MonthlyStatisticsDTO> statsList) {
        if (statsList.isEmpty()) {
            System.out.println("没有找到统计数据");
            return;
        }
        
        System.out.println("所有员工月度排班统计报表 (" + statsList.size() + " 名员工):");
        printSeparator();
        System.out.printf("%-15s %-8s %-8s %-12s %-12s %-12s%n", 
                "员工ID", "工作日", "休息日", "早班", "中班", "夜班");
        printSeparator();
        
        for (MonthlyStatisticsDTO stats : statsList) {
            System.out.printf("%-15s %-8s %-8s %-12s %-12s %-12s%n",
                    stats.getEmployeeId(),
                    stats.getTotalWorkingDays(),
                    stats.getTotalRestDays(),
                    stats.getMorningShiftCount(),
                    stats.getAfternoonShiftCount(),
                    stats.getNightShiftCount());
        }
        printSeparator();
    }

    private static String formatShiftType(ShiftType type) {
        if (type == null) {
            return "N/A";
        }
        return type.getDescription();
    }

    private static String padRight(String s, int n) {
        return String.format("%-" + n + "s", s);
    }
}