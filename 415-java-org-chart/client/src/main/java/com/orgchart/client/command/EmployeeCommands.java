package com.orgchart.client.command;

import com.fasterxml.jackson.core.type.TypeReference;
import com.orgchart.common.dto.ApiResponse;
import com.orgchart.common.dto.EmployeeDTO;
import com.orgchart.common.dto.ReportingLineDTO;
import com.orgchart.common.dto.TransferHistoryDTO;
import com.orgchart.common.dto.request.CreateEmployeeRequest;
import com.orgchart.common.dto.request.TransferEmployeeRequest;
import com.orgchart.common.dto.request.UpdateEmployeeRequest;
import com.orgchart.client.http.HttpClient;
import com.orgchart.client.util.OutputFormatter;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public class EmployeeCommands implements Command {

    @Override
    public String getName() {
        return "employee";
    }

    @Override
    public String getDescription() {
        return "员工管理命令 (list, get, create, update, delete, transfer, reporting-line, transfer-history)";
    }

    @Override
    public void execute(String[] args, HttpClient httpClient) throws Exception {
        OutputFormatter formatter = new OutputFormatter(httpClient.getObjectMapper());

        if (args.length == 0) {
            printHelp();
            return;
        }

        String subCommand = args[0];
        String[] remainingArgs = args.length > 1 ? Arrays.copyOfRange(args, 1, args.length) : new String[0];

        switch (subCommand) {
            case "list":
                listEmployees(httpClient, formatter);
                break;
            case "get":
                getEmployee(remainingArgs, httpClient, formatter);
                break;
            case "create":
                createEmployee(remainingArgs, httpClient, formatter);
                break;
            case "update":
                updateEmployee(remainingArgs, httpClient, formatter);
                break;
            case "delete":
                deleteEmployee(remainingArgs, httpClient, formatter);
                break;
            case "transfer":
                transferEmployee(remainingArgs, httpClient, formatter);
                break;
            case "reporting-line":
                getReportingLine(remainingArgs, httpClient, formatter);
                break;
            case "transfer-history":
                getTransferHistory(remainingArgs, httpClient, formatter);
                break;
            default:
                System.out.println("未知的子命令: " + subCommand);
                printHelp();
        }
    }

    private void listEmployees(HttpClient httpClient, OutputFormatter formatter) throws Exception {
        String path = "/api/employees";
        TypeReference<ApiResponse<List<EmployeeDTO>>> typeRef = new TypeReference<ApiResponse<List<EmployeeDTO>>>() {};
        ApiResponse<List<EmployeeDTO>> response = httpClient.get(path, typeRef.getType());

        if (response.getCode() == 0) {
            List<EmployeeDTO> employees = response.getData();
            if (employees == null || employees.isEmpty()) {
                System.out.println("暂无员工数据");
                return;
            }

            System.out.println("员工列表:");
            System.out.println();

            List<String> headers = Arrays.asList("ID", "姓名", "邮箱", "部门", "上级");
            List<List<String>> rows = new ArrayList<>();
            for (EmployeeDTO emp : employees) {
                List<String> row = new ArrayList<>();
                row.add(emp.getId());
                row.add(emp.getName());
                row.add(emp.getEmail() != null ? emp.getEmail() : "-");
                row.add(emp.getDepartmentName() != null ? emp.getDepartmentName() : "-");
                row.add(emp.getManagerName() != null ? emp.getManagerName() : "-");
                rows.add(row);
            }
            formatter.printTable(headers, rows);
        } else {
            formatter.printResult(response);
        }
    }

    private void getEmployee(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length == 0) {
            System.out.println("用法: employee get <id>");
            return;
        }

        String id = args[0];
        String path = "/api/employees/" + id;
        TypeReference<ApiResponse<EmployeeDTO>> typeRef = new TypeReference<ApiResponse<EmployeeDTO>>() {};
        ApiResponse<EmployeeDTO> response = httpClient.get(path, typeRef.getType());

        if (response.getCode() == 0) {
            EmployeeDTO emp = response.getData();
            System.out.println("员工详情:");
            System.out.println();

            Map<String, String> map = new LinkedHashMap<>();
            map.put("ID", emp.getId());
            map.put("姓名", emp.getName());
            map.put("邮箱", emp.getEmail() != null ? emp.getEmail() : "-");
            map.put("电话", emp.getPhone() != null ? emp.getPhone() : "-");
            map.put("部门ID", emp.getDepartmentId() != null ? emp.getDepartmentId() : "无");
            map.put("部门名称", emp.getDepartmentName() != null ? emp.getDepartmentName() : "无");
            map.put("上级ID", emp.getManagerId() != null ? emp.getManagerId() : "无");
            map.put("上级名称", emp.getManagerName() != null ? emp.getManagerName() : "无");
            
            String virtualTeams = emp.getVirtualTeamNames() != null && !emp.getVirtualTeamNames().isEmpty()
                    ? String.join(", ", emp.getVirtualTeamNames()) : "无";
            map.put("虚拟团队", virtualTeams);
            
            map.put("创建时间", emp.getCreatedAt() != null ? emp.getCreatedAt().toString() : "-");
            map.put("更新时间", emp.getUpdatedAt() != null ? emp.getUpdatedAt().toString() : "-");

            formatter.printKeyValue(map);
        } else {
            formatter.printResult(response);
        }
    }

    private void createEmployee(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        Map<String, String> params = parseParams(args);

        if (!params.containsKey("name") || !params.containsKey("departmentId")) {
            System.out.println("用法: employee create --name <姓名> --departmentId <部门ID> [--email <邮箱>] [--phone <电话>] [--managerId <上级ID>]");
            return;
        }

        CreateEmployeeRequest request = new CreateEmployeeRequest();
        request.setName(params.get("name"));
        request.setDepartmentId(params.get("departmentId"));
        request.setEmail(params.get("email"));
        request.setPhone(params.get("phone"));
        request.setManagerId(params.get("managerId"));

        String path = "/api/employees";
        TypeReference<ApiResponse<EmployeeDTO>> typeRef = new TypeReference<ApiResponse<EmployeeDTO>>() {};
        ApiResponse<EmployeeDTO> response = httpClient.post(path, request, typeRef.getType());
        formatter.printResult(response);
    }

    private void updateEmployee(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length == 0) {
            System.out.println("用法: employee update <id> [--name <姓名>] [--email <邮箱>] [--phone <电话>] [--departmentId <部门ID>] [--managerId <上级ID>]");
            return;
        }

        String id = args[0];
        Map<String, String> params = parseParams(Arrays.copyOfRange(args, 1, args.length));

        if (params.isEmpty()) {
            System.out.println("至少需要指定一个更新字段");
            return;
        }

        UpdateEmployeeRequest request = new UpdateEmployeeRequest();
        request.setName(params.get("name"));
        request.setEmail(params.get("email"));
        request.setPhone(params.get("phone"));
        request.setDepartmentId(params.get("departmentId"));
        request.setManagerId(params.get("managerId"));

        String path = "/api/employees/" + id;
        TypeReference<ApiResponse<EmployeeDTO>> typeRef = new TypeReference<ApiResponse<EmployeeDTO>>() {};
        ApiResponse<EmployeeDTO> response = httpClient.put(path, request, typeRef.getType());
        formatter.printResult(response);
    }

    private void deleteEmployee(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length == 0) {
            System.out.println("用法: employee delete <id>");
            return;
        }

        String id = args[0];
        String path = "/api/employees/" + id;
        TypeReference<ApiResponse<Void>> typeRef = new TypeReference<ApiResponse<Void>>() {};
        ApiResponse<Void> response = httpClient.delete(path, typeRef.getType());
        formatter.printResult(response);
    }

    private void transferEmployee(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length < 2) {
            System.out.println("用法: employee transfer <员工ID> --departmentId <新部门ID> [--reason <原因>]");
            return;
        }

        String id = args[0];
        Map<String, String> params = parseParams(Arrays.copyOfRange(args, 1, args.length));

        if (!params.containsKey("departmentId")) {
            System.out.println("请指定目标部门ID: --departmentId <新部门ID>");
            return;
        }

        TransferEmployeeRequest request = new TransferEmployeeRequest();
        request.setNewDepartmentId(params.get("departmentId"));
        request.setReason(params.get("reason"));

        String path = "/api/employees/" + id + "/transfer";
        TypeReference<ApiResponse<EmployeeDTO>> typeRef = new TypeReference<ApiResponse<EmployeeDTO>>() {};
        ApiResponse<EmployeeDTO> response = httpClient.post(path, request, typeRef.getType());
        formatter.printResult(response);
    }

    private void getReportingLine(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length == 0) {
            System.out.println("用法: employee reporting-line <员工ID>");
            return;
        }

        String id = args[0];
        String path = "/api/employees/" + id + "/reporting-line";
        TypeReference<ApiResponse<ReportingLineDTO>> typeRef = new TypeReference<ApiResponse<ReportingLineDTO>>() {};
        ApiResponse<ReportingLineDTO> response = httpClient.get(path, typeRef.getType());

        if (response.getCode() == 0) {
            ReportingLineDTO reportingLine = response.getData();
            System.out.println("汇报线:");
            System.out.println();
            System.out.println("员工: " + reportingLine.getEmployeeName() + " (" + reportingLine.getEmployeeId() + ")");
            System.out.println();

            if (reportingLine.getManagers() != null && !reportingLine.getManagers().isEmpty()) {
                System.out.println("上级汇报线:");
                for (ReportingLineDTO.ManagerNode manager : reportingLine.getManagers()) {
                    System.out.println("  第" + manager.getLevel() + "级: " + manager.getName() + 
                            " (" + manager.getId() + ") - " + 
                            (manager.getDepartmentName() != null ? manager.getDepartmentName() : ""));
                }
            } else {
                System.out.println("该员工没有上级汇报人");
            }

            System.out.println();
            if (reportingLine.getDepartmentPath() != null && !reportingLine.getDepartmentPath().isEmpty()) {
                System.out.println("部门路径:");
                for (ReportingLineDTO.DepartmentNode dept : reportingLine.getDepartmentPath()) {
                    System.out.println("  " + dept.getName() + " (" + dept.getId() + ")");
                }
            }
        } else {
            formatter.printResult(response);
        }
    }

    private void getTransferHistory(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length == 0) {
            System.out.println("用法: employee transfer-history <员工ID>");
            return;
        }

        String id = args[0];
        String path = "/api/employees/" + id + "/transfer-history";
        TypeReference<ApiResponse<List<TransferHistoryDTO>>> typeRef = new TypeReference<ApiResponse<List<TransferHistoryDTO>>>() {};
        ApiResponse<List<TransferHistoryDTO>> response = httpClient.get(path, typeRef.getType());

        if (response.getCode() == 0) {
            List<TransferHistoryDTO> history = response.getData();
            if (history == null || history.isEmpty()) {
                System.out.println("该员工暂无调动记录");
                return;
            }

            System.out.println("调动历史:");
            System.out.println();

            List<String> headers = Arrays.asList("时间", "原部门", "目标部门", "原因");
            List<List<String>> rows = new ArrayList<>();
            for (TransferHistoryDTO h : history) {
                List<String> row = new ArrayList<>();
                row.add(h.getTransferTime() != null ? h.getTransferTime().toString() : "-");
                row.add(h.getFromDepartmentName() != null ? h.getFromDepartmentName() : "-");
                row.add(h.getToDepartmentName() != null ? h.getToDepartmentName() : "-");
                row.add(h.getReason() != null ? h.getReason() : "-");
                rows.add(row);
            }
            formatter.printTable(headers, rows);
        } else {
            formatter.printResult(response);
        }
    }

    private Map<String, String> parseParams(String[] args) {
        Map<String, String> params = new HashMap<>();
        for (int i = 0; i < args.length; i++) {
            String arg = args[i];
            if (arg.startsWith("--") && i + 1 < args.length) {
                String key = arg.substring(2);
                String value = args[i + 1];
                if (!value.startsWith("--")) {
                    params.put(key, value);
                    i++;
                }
            }
        }
        return params;
    }

    private void printHelp() {
        System.out.println();
        System.out.println("员工管理命令:");
        System.out.println();
        System.out.println("  employee list                     列出所有员工");
        System.out.println("  employee get <id>                获取员工详情");
        System.out.println("  employee create --name <姓名> --departmentId <部门ID>");
        System.out.println("                                    创建员工");
        System.out.println("  employee update <id> [--name <姓名>] [--email <邮箱>] ...");
        System.out.println("                                    更新员工");
        System.out.println("  employee delete <id>              删除员工");
        System.out.println("  employee transfer <员工ID> --departmentId <新部门ID>");
        System.out.println("                                    员工调部门");
        System.out.println("  employee reporting-line <员工ID>");
        System.out.println("                                    查询汇报线");
        System.out.println("  employee transfer-history <员工ID>");
        System.out.println("                                    查询调动历史");
        System.out.println();
    }
}
