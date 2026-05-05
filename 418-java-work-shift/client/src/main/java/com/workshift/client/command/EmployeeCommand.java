package com.workshift.client.command;

import com.fasterxml.jackson.core.type.TypeReference;
import com.workshift.common.dto.EmployeeDTO;
import com.workshift.common.response.ApiResponse;
import com.workshift.client.http.HttpClient;
import com.workshift.client.util.OutputFormatter;
import java.math.BigDecimal;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class EmployeeCommand implements Command {

    private final HttpClient httpClient;

    public EmployeeCommand(HttpClient httpClient) {
        this.httpClient = httpClient;
    }

    @Override
    public void execute(String[] args) throws Exception {
        if (args.length < 2) {
            printHelp();
            return;
        }

        String action = args[1];

        switch (action.toLowerCase()) {
            case "list":
                listEmployees();
                break;
            case "get":
                if (args.length < 3) {
                    System.out.println("用法: employee get <id>");
                    return;
                }
                getEmployee(args[2]);
                break;
            case "create":
                if (args.length < 5) {
                    System.out.println("用法: employee create <name> <department> <hourlyWage>");
                    return;
                }
                createEmployee(args[2], args[3], args[4]);
                break;
            case "update":
                if (args.length < 3) {
                    System.out.println("用法: employee update <id> [name=<name>] [dept=<dept>] [wage=<wage>]");
                    return;
                }
                updateEmployee(args[2], args);
                break;
            case "delete":
                if (args.length < 3) {
                    System.out.println("用法: employee delete <id>");
                    return;
                }
                deleteEmployee(args[2]);
                break;
            default:
                System.out.println("未知操作: " + action);
                printHelp();
        }
    }

    private void listEmployees() throws Exception {
        TypeReference<ApiResponse<List<EmployeeDTO>>> typeRef = 
                new TypeReference<ApiResponse<List<EmployeeDTO>>>() {};
        ApiResponse<List<EmployeeDTO>> response = httpClient.get("/api/employees", typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printEmployees(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void getEmployee(String id) throws Exception {
        TypeReference<ApiResponse<EmployeeDTO>> typeRef = 
                new TypeReference<ApiResponse<EmployeeDTO>>() {};
        ApiResponse<EmployeeDTO> response = httpClient.get("/api/employees/" + id, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printEmployee(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void createEmployee(String name, String department, String wageStr) throws Exception {
        BigDecimal wage;
        try {
            wage = new BigDecimal(wageStr);
        } catch (NumberFormatException e) {
            OutputFormatter.printError("无效的时薪: " + wageStr);
            return;
        }

        EmployeeDTO employee = new EmployeeDTO();
        employee.setName(name);
        employee.setDepartment(department);
        employee.setHourlyWage(wage);

        TypeReference<ApiResponse<EmployeeDTO>> typeRef = 
                new TypeReference<ApiResponse<EmployeeDTO>>() {};
        ApiResponse<EmployeeDTO> response = httpClient.post("/api/employees", employee, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSuccess("员工创建成功");
            OutputFormatter.printEmployee(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void updateEmployee(String id, String[] args) throws Exception {
        Map<String, String> updates = parseUpdates(args, 3);
        
        EmployeeDTO employee = new EmployeeDTO();
        if (updates.containsKey("name")) {
            employee.setName(updates.get("name"));
        }
        if (updates.containsKey("dept")) {
            employee.setDepartment(updates.get("dept"));
        }
        if (updates.containsKey("wage")) {
            try {
                employee.setHourlyWage(new BigDecimal(updates.get("wage")));
            } catch (NumberFormatException e) {
                OutputFormatter.printError("无效的时薪: " + updates.get("wage"));
                return;
            }
        }

        TypeReference<ApiResponse<Void>> typeRef = 
                new TypeReference<ApiResponse<Void>>() {};
        ApiResponse<Void> response = httpClient.put("/api/employees/" + id, employee, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSuccess("员工更新成功");
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void deleteEmployee(String id) throws Exception {
        TypeReference<ApiResponse<Void>> typeRef = 
                new TypeReference<ApiResponse<Void>>() {};
        ApiResponse<Void> response = httpClient.delete("/api/employees/" + id, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSuccess("员工删除成功");
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private Map<String, String> parseUpdates(String[] args, int startIndex) {
        Map<String, String> updates = new HashMap<>();
        for (int i = startIndex; i < args.length; i++) {
            String arg = args[i];
            int eqIndex = arg.indexOf('=');
            if (eqIndex > 0) {
                String key = arg.substring(0, eqIndex);
                String value = arg.substring(eqIndex + 1);
                updates.put(key, value);
            }
        }
        return updates;
    }

    private void printHelp() {
        System.out.println("员工管理命令:");
        System.out.println("  list                              - 列出所有员工");
        System.out.println("  get <id>                          - 获取指定员工信息");
        System.out.println("  create <name> <dept> <wage>      - 创建员工");
        System.out.println("  update <id> [name=...] [dept=...] [wage=...] - 更新员工");
        System.out.println("  delete <id>                       - 删除员工");
    }
}