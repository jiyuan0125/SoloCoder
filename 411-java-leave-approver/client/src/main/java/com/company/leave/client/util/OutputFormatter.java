package com.company.leave.client.util;

import com.company.leave.dto.CalendarViewDTO;
import com.company.leave.enums.LeaveStatus;
import com.company.leave.enums.LeaveType;

import java.time.LocalDate;
import java.util.List;
import java.util.Map;

public class OutputFormatter {

    public static void printEmployee(Map<String, Object> employee) {
        System.out.println("========================================");
        System.out.println("员工信息");
        System.out.println("========================================");
        System.out.printf("ID: %s%n", employee.get("id"));
        System.out.printf("姓名: %s%n", employee.get("name"));
        System.out.printf("上级ID: %s%n", employee.get("managerId"));
        System.out.printf("入职日期: %s%n", employee.get("joinDate"));
        System.out.printf("年假配额: %s天%n", employee.get("annualLeaveQuota"));
        System.out.printf("剩余年假: %s天%n", employee.get("annualLeaveRemaining"));
        System.out.printf("结转年假: %s天%n", employee.get("carriedOverLeave"));
        System.out.println("========================================");
    }

    public static void printEmployeeList(List<Map<String, Object>> employees) {
        System.out.println("========================================");
        System.out.println("员工列表");
        System.out.println("========================================");
        System.out.printf("%-5s %-10s %-10s %-12s %-8s %-8s%n", 
                "ID", "姓名", "上级ID", "入职日期", "年假配额", "剩余年假");
        System.out.println("------------------------------------------------------------");
        for (Map<String, Object> emp : employees) {
            System.out.printf("%-5s %-10s %-10s %-12s %-8s %-8s%n",
                    emp.get("id"),
                    emp.get("name"),
                    emp.get("managerId"),
                    emp.get("joinDate"),
                    emp.get("annualLeaveQuota"),
                    emp.get("annualLeaveRemaining"));
        }
        System.out.println("========================================");
    }

    public static void printLeaveRecord(Map<String, Object> leave) {
        System.out.println("========================================");
        System.out.println("请假记录");
        System.out.println("========================================");
        System.out.printf("ID: %s%n", leave.get("id"));
        System.out.printf("员工ID: %s%n", leave.get("employeeId"));
        System.out.printf("员工姓名: %s%n", leave.get("employeeName"));
        System.out.printf("请假类型: %s%n", formatLeaveType(leave.get("leaveType")));
        System.out.printf("开始日期: %s%n", leave.get("startDate"));
        System.out.printf("结束日期: %s%n", leave.get("endDate"));
        System.out.printf("实际天数: %s天%n", leave.get("actualDays"));
        System.out.printf("事由: %s%n", leave.get("reason"));
        if (leave.get("attachmentName") != null) {
            System.out.printf("附件: %s%n", leave.get("attachmentName"));
        }
        System.out.printf("状态: %s%n", formatLeaveStatus(leave.get("status")));
        if (leave.get("approvedBy") != null) {
            System.out.printf("审批人ID: %s%n", leave.get("approvedBy"));
        }
        if (leave.get("approvalComment") != null) {
            System.out.printf("审批意见: %s%n", leave.get("approvalComment"));
        }
        System.out.printf("创建时间: %s%n", leave.get("createdAt"));
        System.out.println("========================================");
    }

    public static void printLeaveList(List<Map<String, Object>> leaves) {
        System.out.println("========================================");
        System.out.println("请假记录列表");
        System.out.println("========================================");
        System.out.printf("%-5s %-10s %-8s %-12s %-12s %-6s %-10s%n",
                "ID", "员工", "类型", "开始日期", "结束日期", "天数", "状态");
        System.out.println("----------------------------------------------------------------------");
        for (Map<String, Object> leave : leaves) {
            System.out.printf("%-5s %-10s %-8s %-12s %-12s %-6s %-10s%n",
                    leave.get("id"),
                    leave.get("employeeName"),
                    formatLeaveTypeShort(leave.get("leaveType")),
                    leave.get("startDate"),
                    leave.get("endDate"),
                    leave.get("actualDays"),
                    formatLeaveStatus(leave.get("status")));
        }
        System.out.println("========================================");
    }

    public static void printCalendar(CalendarViewDTO calendar) {
        System.out.println("========================================");
        System.out.printf("团队日历 - %d年%d月%n", calendar.getYear(), calendar.getMonth());
        System.out.println("========================================");
        
        List<CalendarViewDTO.DayLeaveStatus> dayStatuses = calendar.getDayStatuses();
        if (dayStatuses.isEmpty()) {
            System.out.println("本月无请假记录");
            System.out.println("========================================");
            return;
        }

        System.out.println("日期          请假人员");
        System.out.println("-----------------------------");
        for (CalendarViewDTO.DayLeaveStatus dayStatus : dayStatuses) {
            List<CalendarViewDTO.EmployeeLeave> leaves = dayStatus.getLeaves();
            if (!leaves.isEmpty()) {
                System.out.printf("%s  ", dayStatus.getDate());
                for (int i = 0; i < leaves.size(); i++) {
                    CalendarViewDTO.EmployeeLeave leave = leaves.get(i);
                    System.out.printf("%s(%s)", 
                            leave.getEmployeeName(), 
                            formatLeaveTypeShort(leave.getLeaveType().name()));
                    if (i < leaves.size() - 1) {
                        System.out.print(", ");
                    }
                }
                System.out.println();
            }
        }
        System.out.println("========================================");
    }

    public static void printSuccess(String message) {
        System.out.println("[成功] " + message);
    }

    public static void printError(String message) {
        System.err.println("[错误] " + message);
    }

    private static String formatLeaveType(Object type) {
        if (type == null) return "";
        String typeStr = type.toString();
        return switch (typeStr) {
            case "ANNUAL" -> "年假";
            case "SICK" -> "病假";
            case "PERSONAL" -> "事假";
            case "MARRIAGE" -> "婚假";
            case "MATERNITY" -> "产假";
            case "PATERNITY" -> "陪产假";
            default -> typeStr;
        };
    }

    private static String formatLeaveTypeShort(Object type) {
        if (type == null) return "";
        String typeStr = type.toString();
        return switch (typeStr) {
            case "ANNUAL" -> "年假";
            case "SICK" -> "病假";
            case "PERSONAL" -> "事假";
            case "MARRIAGE" -> "婚假";
            case "MATERNITY" -> "产假";
            case "PATERNITY" -> "陪产";
            default -> typeStr;
        };
    }

    private static String formatLeaveStatus(Object status) {
        if (status == null) return "";
        String statusStr = status.toString();
        return switch (statusStr) {
            case "PENDING" -> "待审批";
            case "APPROVED" -> "已批准";
            case "REJECTED" -> "已驳回";
            default -> statusStr;
        };
    }
}
