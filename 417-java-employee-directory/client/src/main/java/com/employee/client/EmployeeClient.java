package com.employee.client;

import com.employee.client.http.HttpClientWrapper;
import com.employee.common.dto.*;
import com.employee.common.response.ApiResponse;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;

import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.Scanner;

public class EmployeeClient {
    private static final String BASE_URL = "http://localhost:8080";
    private final HttpClientWrapper httpClient;
    private final ObjectMapper objectMapper;
    private final Scanner scanner;

    public EmployeeClient() {
        this.httpClient = new HttpClientWrapper(BASE_URL);
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
        this.scanner = new Scanner(System.in);
    }

    public static void main(String[] args) {
        EmployeeClient client = new EmployeeClient();
        if (args.length > 0) {
            client.runWithArgs(args);
        } else {
            client.runInteractive();
        }
    }

    private void runInteractive() {
        System.out.println("=== 员工通讯录客户端 ===");
        System.out.println("请选择当前用户角色:");
        System.out.println("1. 普通员工 (EMPLOYEE)");
        System.out.println("2. HR (HR)");
        System.out.println("3. 管理员 (ADMIN)");
        System.out.print("请输入选择 (1-3): ");
        
        int roleChoice = Integer.parseInt(scanner.nextLine());
        switch (roleChoice) {
            case 2 -> httpClient.setCurrentUser("HR001", "HR经理", "HR");
            case 3 -> httpClient.setCurrentUser("ADM001", "系统管理员", "ADMIN");
            default -> httpClient.setCurrentUser("E001", "张三", "EMPLOYEE");
        }

        while (true) {
            System.out.println("\n=== 功能菜单 ===");
            System.out.println("1. 新增员工");
            System.out.println("2. 更新员工信息 (HR/管理员)");
            System.out.println("3. 更新个人信息 (手机号/邮箱)");
            System.out.println("4. 查看员工详情");
            System.out.println("5. 搜索员工");
            System.out.println("6. 查看员工列表");
            System.out.println("7. 批量导入员工");
            System.out.println("8. 部门人数统计");
            System.out.println("9. 组织架构树");
            System.out.println("10. 查看操作日志");
            System.out.println("0. 退出");
            System.out.print("请输入选择: ");

            String choice = scanner.nextLine();
            try {
                switch (choice) {
                    case "1" -> createEmployee();
                    case "2" -> updateEmployee();
                    case "3" -> updateSelfInfo();
                    case "4" -> getEmployeeById();
                    case "5" -> searchEmployees();
                    case "6" -> listEmployees();
                    case "7" -> batchImport();
                    case "8" -> getDepartmentStats();
                    case "9" -> getOrganizationTree();
                    case "10" -> getLogs();
                    case "0" -> {
                        System.out.println("再见!");
                        return;
                    }
                    default -> System.out.println("无效选择");
                }
            } catch (Exception e) {
                System.out.println("错误: " + e.getMessage());
            }
        }
    }

    private void runWithArgs(String[] args) {
        if (args.length < 1) {
            printHelp();
            return;
        }
        System.out.println("命令行模式暂未实现完整功能，请使用交互模式");
    }

    private void printHelp() {
        System.out.println("员工通讯录客户端");
        System.out.println("用法: java -jar client.jar [命令]");
        System.out.println("不带参数启动交互模式");
    }

    private void createEmployee() throws IOException, InterruptedException {
        System.out.println("\n=== 新增员工 ===");
        EmployeeCreateRequest request = new EmployeeCreateRequest();
        
        System.out.print("工号: ");
        request.setEmployeeId(scanner.nextLine());
        
        System.out.print("姓名: ");
        request.setName(scanner.nextLine());
        
        System.out.print("部门: ");
        request.setDepartment(scanner.nextLine());
        
        System.out.print("职位: ");
        request.setPosition(scanner.nextLine());
        
        System.out.print("手机号: ");
        request.setPhone(scanner.nextLine());
        
        System.out.print("邮箱 (必须是 @company.com): ");
        request.setEmail(scanner.nextLine());
        
        System.out.print("办公地点: ");
        request.setOfficeLocation(scanner.nextLine());

        ApiResponse<EmployeeDTO> response = httpClient.post("/api/employees", request, 
                new TypeReference<ApiResponse<EmployeeDTO>>() {});
        
        if (response.getCode() == 0) {
            System.out.println("创建成功!");
            printEmployee(response.getData());
        } else {
            System.out.println("创建失败: " + response.getMessage());
        }
    }

    private void updateEmployee() throws IOException, InterruptedException {
        System.out.println("\n=== 更新员工信息 (HR/管理员) ===");
        System.out.print("请输入要更新的员工工号: ");
        String employeeId = scanner.nextLine();
        
        EmployeeUpdateRequest request = new EmployeeUpdateRequest();
        
        System.out.print("姓名 (留空不修改): ");
        String name = scanner.nextLine();
        if (!name.isEmpty()) request.setName(name);
        
        System.out.print("部门 (留空不修改): ");
        String dept = scanner.nextLine();
        if (!dept.isEmpty()) request.setDepartment(dept);
        
        System.out.print("职位 (留空不修改): ");
        String position = scanner.nextLine();
        if (!position.isEmpty()) request.setPosition(position);
        
        System.out.print("手机号 (留空不修改): ");
        String phone = scanner.nextLine();
        if (!phone.isEmpty()) request.setPhone(phone);
        
        System.out.print("邮箱 (留空不修改, 必须是 @company.com): ");
        String email = scanner.nextLine();
        if (!email.isEmpty()) request.setEmail(email);
        
        System.out.print("办公地点 (留空不修改): ");
        String location = scanner.nextLine();
        if (!location.isEmpty()) request.setOfficeLocation(location);
        
        System.out.print("是否标记离职? (y/n, 留空不修改): ");
        String resignedStr = scanner.nextLine();
        if (resignedStr.equals("y")) {
            request.setResigned(true);
        } else if (resignedStr.equals("n")) {
            request.setResigned(false);
        }

        ApiResponse<EmployeeDTO> response = httpClient.put("/api/employees/" + employeeId, request,
                new TypeReference<ApiResponse<EmployeeDTO>>() {});
        
        if (response.getCode() == 0) {
            System.out.println("更新成功!");
            printEmployee(response.getData());
        } else {
            System.out.println("更新失败: " + response.getMessage());
        }
    }

    private void updateSelfInfo() throws IOException, InterruptedException {
        System.out.println("\n=== 更新个人信息 ===");
        EmployeeSelfUpdateRequest request = new EmployeeSelfUpdateRequest();
        
        System.out.print("新手机号 (留空不修改): ");
        String phone = scanner.nextLine();
        if (!phone.isEmpty()) request.setPhone(phone);
        
        System.out.print("新邮箱 (留空不修改, 必须是 @company.com): ");
        String email = scanner.nextLine();
        if (!email.isEmpty()) request.setEmail(email);

        ApiResponse<EmployeeDTO> response = httpClient.put("/api/employees/self", request,
                new TypeReference<ApiResponse<EmployeeDTO>>() {});
        
        if (response.getCode() == 0) {
            System.out.println("更新成功!");
            printEmployee(response.getData());
        } else {
            System.out.println("更新失败: " + response.getMessage());
        }
    }

    private void getEmployeeById() throws IOException, InterruptedException {
        System.out.println("\n=== 查看员工详情 ===");
        System.out.print("请输入员工工号: ");
        String employeeId = scanner.nextLine();
        
        System.out.print("是否包含离职员工? (y/n): ");
        boolean includeResigned = scanner.nextLine().equals("y");

        String url = "/api/employees/" + employeeId + "?includeResigned=" + includeResigned;
        ApiResponse<EmployeeDTO> response = httpClient.get(url,
                new TypeReference<ApiResponse<EmployeeDTO>>() {});
        
        if (response.getCode() == 0) {
            printEmployee(response.getData());
        } else {
            System.out.println("查询失败: " + response.getMessage());
        }
    }

    private void searchEmployees() throws IOException, InterruptedException {
        System.out.println("\n=== 搜索员工 ===");
        SearchRequest request = new SearchRequest();
        
        System.out.print("搜索关键词 (姓名/部门/职位): ");
        request.setKeyword(scanner.nextLine());
        
        System.out.print("是否包含离职员工? (y/n): ");
        request.setIncludeResigned(scanner.nextLine().equals("y"));

        ApiResponse<List<EmployeeDTO>> response = httpClient.post("/api/employees/search", request,
                new TypeReference<ApiResponse<List<EmployeeDTO>>>() {});
        
        if (response.getCode() == 0) {
            List<EmployeeDTO> employees = response.getData();
            System.out.println("搜索结果 (共" + employees.size() + "人):");
            for (EmployeeDTO emp : employees) {
                printEmployeeSummary(emp);
            }
        } else {
            System.out.println("搜索失败: " + response.getMessage());
        }
    }

    private void listEmployees() throws IOException, InterruptedException {
        System.out.println("\n=== 员工列表 ===");
        System.out.print("是否包含离职员工? (y/n): ");
        boolean includeResigned = scanner.nextLine().equals("y");

        String url = "/api/employees?includeResigned=" + includeResigned;
        ApiResponse<List<EmployeeDTO>> response = httpClient.get(url,
                new TypeReference<ApiResponse<List<EmployeeDTO>>>() {});
        
        if (response.getCode() == 0) {
            List<EmployeeDTO> employees = response.getData();
            System.out.println("员工列表 (共" + employees.size() + "人):");
            for (EmployeeDTO emp : employees) {
                printEmployeeSummary(emp);
            }
        } else {
            System.out.println("查询失败: " + response.getMessage());
        }
    }

    private void batchImport() throws IOException, InterruptedException {
        System.out.println("\n=== 批量导入员工 ===");
        List<EmployeeCreateRequest> employees = new ArrayList<>();
        
        while (true) {
            System.out.println("\n输入员工信息 (输入空行结束):");
            System.out.print("工号 (空行结束): ");
            String id = scanner.nextLine();
            if (id.isEmpty()) break;
            
            EmployeeCreateRequest request = new EmployeeCreateRequest();
            request.setEmployeeId(id);
            
            System.out.print("姓名: ");
            request.setName(scanner.nextLine());
            
            System.out.print("部门: ");
            request.setDepartment(scanner.nextLine());
            
            System.out.print("职位: ");
            request.setPosition(scanner.nextLine());
            
            System.out.print("手机号: ");
            request.setPhone(scanner.nextLine());
            
            System.out.print("邮箱 (@company.com): ");
            request.setEmail(scanner.nextLine());
            
            System.out.print("办公地点: ");
            request.setOfficeLocation(scanner.nextLine());
            
            employees.add(request);
        }

        if (employees.isEmpty()) {
            System.out.println("未输入任何员工信息");
            return;
        }

        ApiResponse<BatchImportResultDTO> response = httpClient.post("/api/employees/batch-import", employees,
                new TypeReference<ApiResponse<BatchImportResultDTO>>() {});
        
        if (response.getCode() == 0) {
            BatchImportResultDTO result = response.getData();
            System.out.println("\n导入结果:");
            System.out.println("总计: " + result.getTotalCount());
            System.out.println("成功: " + result.getSuccessCount());
            System.out.println("失败: " + result.getErrors().size());
            
            for (ImportErrorDTO error : result.getErrors()) {
                System.out.println("  - 第" + (error.getIndex() + 1) + "条, 工号: " 
                        + error.getEmployeeId() + ", 原因: " + error.getReason());
            }
        } else {
            System.out.println("导入失败: " + response.getMessage());
        }
    }

    private void getDepartmentStats() throws IOException, InterruptedException {
        System.out.println("\n=== 部门人数统计 ===");
        ApiResponse<List<DepartmentStatsDTO>> response = httpClient.get("/api/employees/stats/departments",
                new TypeReference<ApiResponse<List<DepartmentStatsDTO>>>() {});
        
        if (response.getCode() == 0) {
            List<DepartmentStatsDTO> stats = response.getData();
            System.out.println("部门统计:");
            for (DepartmentStatsDTO stat : stats) {
                System.out.println("  " + stat.getDepartment() + ":");
                System.out.println("    总计: " + stat.getTotalCount());
                System.out.println("    在职: " + stat.getActiveCount());
                System.out.println("    离职: " + stat.getResignedCount());
            }
        } else {
            System.out.println("查询失败: " + response.getMessage());
        }
    }

    private void getOrganizationTree() throws IOException, InterruptedException {
        System.out.println("\n=== 组织架构树 ===");
        ApiResponse<OrgTreeNodeDTO> response = httpClient.get("/api/employees/organization-tree",
                new TypeReference<ApiResponse<OrgTreeNodeDTO>>() {});
        
        if (response.getCode() == 0) {
            OrgTreeNodeDTO tree = response.getData();
            printOrgTree(tree, 0);
        } else {
            System.out.println("查询失败: " + response.getMessage());
        }
    }

    private void printOrgTree(OrgTreeNodeDTO node, int level) {
        String indent = "  ".repeat(level);
        String marker = level == 0 ? "├─ " : "│  ";
        System.out.println(indent + marker + node.getName() + " (" + node.getEmployeeCount() + "人)");
        
        for (OrgTreeNodeDTO child : node.getChildren()) {
            printOrgTree(child, level + 1);
        }
    }

    private void getLogs() throws IOException, InterruptedException {
        System.out.println("\n=== 操作日志 ===");
        System.out.print("查看特定员工日志? 输入工号 (直接回车查看所有): ");
        String employeeId = scanner.nextLine();
        
        ApiResponse<List<OperationLogDTO>> response;
        if (employeeId.isEmpty()) {
            response = httpClient.get("/api/employees/logs",
                    new TypeReference<ApiResponse<List<OperationLogDTO>>>() {});
        } else {
            response = httpClient.get("/api/employees/" + employeeId + "/logs",
                    new TypeReference<ApiResponse<List<OperationLogDTO>>>() {});
        }
        
        if (response.getCode() == 0) {
            List<OperationLogDTO> logs = response.getData();
            System.out.println("操作日志 (共" + logs.size() + "条):");
            for (OperationLogDTO log : logs) {
                System.out.println("  [" + log.getOperationTime() + "] " 
                        + log.getOperatorName() + "(" + log.getOperatorId() + ") "
                        + "修改了 " + log.getTargetEmployeeName() + " 的 " + log.getFieldName()
                        + ": " + log.getOldValue() + " -> " + log.getNewValue());
            }
        } else {
            System.out.println("查询失败: " + response.getMessage());
        }
    }

    private void printEmployee(EmployeeDTO emp) {
        System.out.println("\n员工详情:");
        System.out.println("  工号: " + emp.getEmployeeId());
        System.out.println("  姓名: " + emp.getName());
        System.out.println("  部门: " + emp.getDepartment());
        System.out.println("  职位: " + emp.getPosition());
        System.out.println("  手机号: " + emp.getPhone());
        System.out.println("  邮箱: " + emp.getEmail());
        System.out.println("  办公地点: " + emp.getOfficeLocation());
        System.out.println("  状态: " + (emp.isResigned() ? "已离职" : "在职"));
        System.out.println("  创建时间: " + emp.getCreatedAt());
        System.out.println("  更新时间: " + emp.getUpdatedAt());
    }

    private void printEmployeeSummary(EmployeeDTO emp) {
        String status = emp.isResigned() ? "[离职]" : "[在职]";
        System.out.println("  " + emp.getEmployeeId() + " | " + emp.getName() 
                + " | " + emp.getDepartment() + " | " + emp.getPosition() 
                + " | " + emp.getPhone() + " | " + emp.getEmail() + " " + status);
    }
}
