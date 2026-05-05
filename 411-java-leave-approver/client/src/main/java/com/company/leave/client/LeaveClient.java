package com.company.leave.client;

import com.company.leave.client.http.HttpClient;
import com.company.leave.client.service.LeaveApiService;
import com.company.leave.client.util.OutputFormatter;
import com.company.leave.dto.*;
import com.company.leave.enums.LeaveType;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.time.LocalDate;
import java.util.List;
import java.util.Map;

public class LeaveClient {
    private static final String DEFAULT_BASE_URL = "http://localhost:8080";
    private final LeaveApiService apiService;
    private final ObjectMapper objectMapper;

    public LeaveClient(String baseUrl) {
        HttpClient httpClient = new HttpClient(baseUrl);
        this.apiService = new LeaveApiService(httpClient);
        this.objectMapper = httpClient.getObjectMapper();
    }

    public static void main(String[] args) {
        if (args.length == 0) {
            printUsage();
            System.exit(1);
        }

        String baseUrl = System.getenv("LEAVE_SERVER_URL");
        if (baseUrl == null || baseUrl.isEmpty()) {
            baseUrl = DEFAULT_BASE_URL;
        }

        LeaveClient client = new LeaveClient(baseUrl);

        try {
            client.executeCommand(args);
        } catch (Exception e) {
            OutputFormatter.printError("执行失败: " + e.getMessage());
            e.printStackTrace();
            System.exit(1);
        }
    }

    private void executeCommand(String[] args) throws Exception {
        String command = args[0];

        switch (command) {
            case "create-employee":
                handleCreateEmployee(args);
                break;
            case "get-employee":
                handleGetEmployee(args);
                break;
            case "list-employees":
                handleListEmployees();
                break;
            case "list-team":
                handleListTeam(args);
                break;
            case "apply-leave":
                handleApplyLeave(args);
                break;
            case "approve-leave":
                handleApproveLeave(args);
                break;
            case "get-leave":
                handleGetLeave(args);
                break;
            case "list-leaves":
                handleListLeaves(args);
                break;
            case "list-all-leaves":
                handleListAllLeaves();
                break;
            case "calendar":
                handleCalendar(args);
                break;
            case "help":
            default:
                printUsage();
        }
    }

    private void handleCreateEmployee(String[] args) throws Exception {
        if (args.length < 3) {
            System.err.println("用法: create-employee <姓名> <上级ID> [入职日期(YYYY-MM-DD)]");
            System.exit(1);
        }

        String name = args[1];
        Long managerId = parseLong(args[2], "上级ID");
        String joinDate = args.length > 3 ? args[3] : LocalDate.now().toString();

        ApiResponse<EmployeeDTO> response = apiService.createEmployee(name, managerId, joinDate);
        if (response.getCode() != 0) {
            OutputFormatter.printError(response.getMessage());
            return;
        }

        Map<String, Object> empData = convertToMap(response.getData());
        OutputFormatter.printSuccess("员工创建成功");
        OutputFormatter.printEmployee(empData);
    }

    private void handleGetEmployee(String[] args) throws Exception {
        if (args.length < 2) {
            System.err.println("用法: get-employee <员工ID>");
            System.exit(1);
        }

        Long id = parseLong(args[1], "员工ID");
        ApiResponse<EmployeeDTO> response = apiService.getEmployee(id);
        
        if (response.getCode() != 0) {
            OutputFormatter.printError(response.getMessage());
            return;
        }

        Map<String, Object> empData = convertToMap(response.getData());
        OutputFormatter.printEmployee(empData);
    }

    private void handleListEmployees() throws Exception {
        ApiResponse<List> response = apiService.getAllEmployees();
        if (response.getCode() != 0) {
            OutputFormatter.printError(response.getMessage());
            return;
        }

        List<Map<String, Object>> employees = (List<Map<String, Object>>) response.getData();
        OutputFormatter.printEmployeeList(employees);
    }

    private void handleListTeam(String[] args) throws Exception {
        if (args.length < 2) {
            System.err.println("用法: list-team <上级ID>");
            System.exit(1);
        }

        Long managerId = parseLong(args[1], "上级ID");
        ApiResponse<List> response = apiService.getTeamMembers(managerId);
        
        if (response.getCode() != 0) {
            OutputFormatter.printError(response.getMessage());
            return;
        }

        List<Map<String, Object>> employees = (List<Map<String, Object>>) response.getData();
        System.out.println("上级ID: " + managerId + " 的团队成员:");
        OutputFormatter.printEmployeeList(employees);
    }

    private void handleApplyLeave(String[] args) throws Exception {
        if (args.length < 6) {
            System.err.println("用法: apply-leave <员工ID> <类型> <开始日期> <结束日期> <事由> [附件名]");
            System.err.println("类型: ANNUAL(年假), SICK(病假), PERSONAL(事假), MARRIAGE(婚假), MATERNITY(产假), PATERNITY(陪产假)");
            System.exit(1);
        }

        Long employeeId = parseLong(args[1], "员工ID");
        LeaveType leaveType = parseLeaveType(args[2]);
        LocalDate startDate = parseDate(args[3], "开始日期");
        LocalDate endDate = parseDate(args[4], "结束日期");
        String reason = args[5];
        String attachmentName = args.length > 6 ? args[6] : null;

        CreateLeaveRequest request = new CreateLeaveRequest();
        request.setEmployeeId(employeeId);
        request.setLeaveType(leaveType);
        request.setStartDate(startDate);
        request.setEndDate(endDate);
        request.setReason(reason);
        request.setAttachmentName(attachmentName);

        ApiResponse<LeaveRecordDTO> response = apiService.createLeave(request);
        
        if (response.getCode() != 0) {
            OutputFormatter.printError(response.getMessage());
            return;
        }

        Map<String, Object> leaveData = convertToMap(response.getData());
        OutputFormatter.printSuccess("请假申请提交成功");
        OutputFormatter.printLeaveRecord(leaveData);
    }

    private void handleApproveLeave(String[] args) throws Exception {
        if (args.length < 4) {
            System.err.println("用法: approve-leave <请假ID> <审批人ID> <approve|reject> [审批意见]");
            System.exit(1);
        }

        Long leaveId = parseLong(args[1], "请假ID");
        Long managerId = parseLong(args[2], "审批人ID");
        boolean approved = args[3].equalsIgnoreCase("approve");
        String comment = args.length > 4 ? args[4] : null;

        ApiResponse<LeaveRecordDTO> response = apiService.approveLeave(leaveId, managerId, approved, comment);
        
        if (response.getCode() != 0) {
            OutputFormatter.printError(response.getMessage());
            return;
        }

        Map<String, Object> leaveData = convertToMap(response.getData());
        OutputFormatter.printSuccess(approved ? "请假已批准" : "请假已驳回");
        OutputFormatter.printLeaveRecord(leaveData);
    }

    private void handleGetLeave(String[] args) throws Exception {
        if (args.length < 2) {
            System.err.println("用法: get-leave <请假ID>");
            System.exit(1);
        }

        Long id = parseLong(args[1], "请假ID");
        ApiResponse<LeaveRecordDTO> response = apiService.getLeave(id);
        
        if (response.getCode() != 0) {
            OutputFormatter.printError(response.getMessage());
            return;
        }

        Map<String, Object> leaveData = convertToMap(response.getData());
        OutputFormatter.printLeaveRecord(leaveData);
    }

    private void handleListLeaves(String[] args) throws Exception {
        if (args.length < 2) {
            System.err.println("用法: list-leaves <员工ID>");
            System.exit(1);
        }

        Long employeeId = parseLong(args[1], "员工ID");
        ApiResponse<List> response = apiService.getEmployeeLeaves(employeeId);
        
        if (response.getCode() != 0) {
            OutputFormatter.printError(response.getMessage());
            return;
        }

        List<Map<String, Object>> leaves = (List<Map<String, Object>>) response.getData();
        System.out.println("员工ID: " + employeeId + " 的请假记录:");
        OutputFormatter.printLeaveList(leaves);
    }

    private void handleListAllLeaves() throws Exception {
        ApiResponse<List> response = apiService.getAllLeaves();
        if (response.getCode() != 0) {
            OutputFormatter.printError(response.getMessage());
            return;
        }

        List<Map<String, Object>> leaves = (List<Map<String, Object>>) response.getData();
        OutputFormatter.printLeaveList(leaves);
    }

    private void handleCalendar(String[] args) throws Exception {
        if (args.length < 4) {
            System.err.println("用法: calendar <上级ID> <年份> <月份>");
            System.exit(1);
        }

        Long managerId = parseLong(args[1], "上级ID");
        int year = parseInt(args[2], "年份");
        int month = parseInt(args[3], "月份");

        ApiResponse<CalendarViewDTO> response = apiService.getTeamCalendar(managerId, year, month);
        
        if (response.getCode() != 0) {
            OutputFormatter.printError(response.getMessage());
            return;
        }

        CalendarViewDTO calendar = objectMapper.convertValue(response.getData(), CalendarViewDTO.class);
        OutputFormatter.printCalendar(calendar);
    }

    private Long parseLong(String value, String fieldName) {
        try {
            return Long.parseLong(value);
        } catch (NumberFormatException e) {
            throw new IllegalArgumentException(fieldName + " 必须是数字: " + value);
        }
    }

    private int parseInt(String value, String fieldName) {
        try {
            return Integer.parseInt(value);
        } catch (NumberFormatException e) {
            throw new IllegalArgumentException(fieldName + " 必须是数字: " + value);
        }
    }

    private LocalDate parseDate(String value, String fieldName) {
        try {
            return LocalDate.parse(value);
        } catch (Exception e) {
            throw new IllegalArgumentException(fieldName + " 格式错误，请使用 YYYY-MM-DD 格式: " + value);
        }
    }

    private LeaveType parseLeaveType(String value) {
        try {
            return LeaveType.valueOf(value.toUpperCase());
        } catch (IllegalArgumentException e) {
            throw new IllegalArgumentException("无效的请假类型: " + value + 
                "。有效类型: ANNUAL, SICK, PERSONAL, MARRIAGE, MATERNITY, PATERNITY");
        }
    }

    private Map<String, Object> convertToMap(Object obj) {
        return objectMapper.convertValue(obj, Map.class);
    }

    private static void printUsage() {
        System.out.println("========================================");
        System.out.println("请假审批系统 - 命令行客户端");
        System.out.println("========================================");
        System.out.println();
        System.out.println("员工管理:");
        System.out.println("  create-employee <姓名> <上级ID> [入职日期]  - 创建新员工");
        System.out.println("  get-employee <员工ID>                        - 查询员工信息");
        System.out.println("  list-employees                                - 列出所有员工");
        System.out.println("  list-team <上级ID>                            - 列出团队成员");
        System.out.println();
        System.out.println("请假管理:");
        System.out.println("  apply-leave <员工ID> <类型> <开始日期> <结束日期> <事由> [附件名]");
        System.out.println("       类型: ANNUAL(年假), SICK(病假), PERSONAL(事假),");
        System.out.println("             MARRIAGE(婚假), MATERNITY(产假), PATERNITY(陪产假)");
        System.out.println("  approve-leave <请假ID> <审批人ID> <approve|reject> [意见]");
        System.out.println("  get-leave <请假ID>                            - 查询请假详情");
        System.out.println("  list-leaves <员工ID>                          - 列出员工请假记录");
        System.out.println("  list-all-leaves                                - 列出所有请假记录");
        System.out.println();
        System.out.println("日历查询:");
        System.out.println("  calendar <上级ID> <年份> <月份>               - 查询团队请假日历");
        System.out.println();
        System.out.println("环境变量:");
        System.out.println("  LEAVE_SERVER_URL  - 服务端地址 (默认: http://localhost:8080)");
        System.out.println();
        System.out.println("示例:");
        System.out.println("  java -jar client.jar create-employee 张三 null 2020-01-01");
        System.out.println("  java -jar client.jar apply-leave 1 ANNUAL 2026-05-10 2026-05-12 \"回家探亲\"");
        System.out.println("  java -jar client.jar approve-leave 1 2 approve \"同意\"");
        System.out.println("  java -jar client.jar calendar 2 2026 5");
    }
}
