package com.orgchart.client.command;

import com.fasterxml.jackson.core.type.TypeReference;
import com.orgchart.common.dto.ApiResponse;
import com.orgchart.common.dto.DepartmentDTO;
import com.orgchart.common.dto.EmployeeDTO;
import com.orgchart.common.dto.request.CreateDepartmentRequest;
import com.orgchart.common.dto.request.MergeDepartmentRequest;
import com.orgchart.common.dto.request.UpdateDepartmentRequest;
import com.orgchart.client.http.HttpClient;
import com.orgchart.client.util.OutputFormatter;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public class DepartmentCommands implements Command {

    @Override
    public String getName() {
        return "department";
    }

    @Override
    public String getDescription() {
        return "部门管理命令 (list, get, create, update, delete, merge, tree)";
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
                listDepartments(httpClient, formatter);
                break;
            case "get":
                getDepartment(remainingArgs, httpClient, formatter);
                break;
            case "create":
                createDepartment(remainingArgs, httpClient, formatter);
                break;
            case "update":
                updateDepartment(remainingArgs, httpClient, formatter);
                break;
            case "delete":
                deleteDepartment(remainingArgs, httpClient, formatter);
                break;
            case "merge":
                mergeDepartment(remainingArgs, httpClient, formatter);
                break;
            case "tree":
                getDepartmentTree(httpClient, formatter);
                break;
            case "employees":
                getDepartmentEmployees(remainingArgs, httpClient, formatter);
                break;
            default:
                System.out.println("未知的子命令: " + subCommand);
                printHelp();
        }
    }

    private void listDepartments(HttpClient httpClient, OutputFormatter formatter) throws Exception {
        String path = "/api/departments";
        TypeReference<ApiResponse<List<DepartmentDTO>>> typeRef = new TypeReference<ApiResponse<List<DepartmentDTO>>>() {};
        ApiResponse<List<DepartmentDTO>> response = httpClient.get(path, typeRef.getType());

        if (response.getCode() == 0) {
            List<DepartmentDTO> departments = response.getData();
            if (departments == null || departments.isEmpty()) {
                System.out.println("暂无部门数据");
                return;
            }

            System.out.println("部门列表:");
            System.out.println();

            List<String> headers = Arrays.asList("ID", "名称", "父部门ID", "层级");
            List<List<String>> rows = new ArrayList<>();
            for (DepartmentDTO dept : departments) {
                List<String> row = new ArrayList<>();
                row.add(dept.getId());
                row.add(dept.getName());
                row.add(dept.getParentId() != null ? dept.getParentId() : "-");
                row.add(String.valueOf(dept.getLevel()));
                rows.add(row);
            }
            formatter.printTable(headers, rows);
        } else {
            formatter.printResult(response);
        }
    }

    private void getDepartment(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length == 0) {
            System.out.println("用法: department get <id>");
            return;
        }

        String id = args[0];
        String path = "/api/departments/" + id;
        TypeReference<ApiResponse<DepartmentDTO>> typeRef = new TypeReference<ApiResponse<DepartmentDTO>>() {};
        ApiResponse<DepartmentDTO> response = httpClient.get(path, typeRef.getType());

        if (response.getCode() == 0) {
            DepartmentDTO dept = response.getData();
            System.out.println("部门详情:");
            System.out.println();

            Map<String, String> map = new LinkedHashMap<>();
            map.put("ID", dept.getId());
            map.put("名称", dept.getName());
            map.put("父部门ID", dept.getParentId() != null ? dept.getParentId() : "无");
            map.put("父部门名称", dept.getParentName() != null ? dept.getParentName() : "无");
            map.put("层级", String.valueOf(dept.getLevel()));
            map.put("创建时间", dept.getCreatedAt() != null ? dept.getCreatedAt().toString() : "-");
            map.put("更新时间", dept.getUpdatedAt() != null ? dept.getUpdatedAt().toString() : "-");
            
            formatter.printKeyValue(map);
        } else {
            formatter.printResult(response);
        }
    }

    private void createDepartment(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        Map<String, String> params = parseParams(args);
        
        if (!params.containsKey("name")) {
            System.out.println("用法: department create --name <名称> [--parentId <父部门ID>]");
            return;
        }

        CreateDepartmentRequest request = new CreateDepartmentRequest();
        request.setName(params.get("name"));
        request.setParentId(params.get("parentId"));

        String path = "/api/departments";
        TypeReference<ApiResponse<DepartmentDTO>> typeRef = new TypeReference<ApiResponse<DepartmentDTO>>() {};
        ApiResponse<DepartmentDTO> response = httpClient.post(path, request, typeRef.getType());
        formatter.printResult(response);
    }

    private void updateDepartment(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length == 0) {
            System.out.println("用法: department update <id> --name <新名称> [--parentId <新父部门ID>]");
            return;
        }

        String id = args[0];
        Map<String, String> params = parseParams(Arrays.copyOfRange(args, 1, args.length));

        if (params.isEmpty()) {
            System.out.println("至少需要指定一个更新字段: --name 或 --parentId");
            return;
        }

        UpdateDepartmentRequest request = new UpdateDepartmentRequest();
        request.setName(params.get("name"));
        request.setParentId(params.get("parentId"));

        String path = "/api/departments/" + id;
        TypeReference<ApiResponse<DepartmentDTO>> typeRef = new TypeReference<ApiResponse<DepartmentDTO>>() {};
        ApiResponse<DepartmentDTO> response = httpClient.put(path, request, typeRef.getType());
        formatter.printResult(response);
    }

    private void deleteDepartment(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length == 0) {
            System.out.println("用法: department delete <id>");
            return;
        }

        String id = args[0];
        String path = "/api/departments/" + id;
        TypeReference<ApiResponse<Void>> typeRef = new TypeReference<ApiResponse<Void>>() {};
        ApiResponse<Void> response = httpClient.delete(path, typeRef.getType());
        formatter.printResult(response);
    }

    private void mergeDepartment(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length < 2) {
            System.out.println("用法: department merge <源部门ID> --targetId <目标部门ID>");
            return;
        }

        String sourceId = args[0];
        Map<String, String> params = parseParams(Arrays.copyOfRange(args, 1, args.length));

        if (!params.containsKey("targetId")) {
            System.out.println("请指定目标部门ID: --targetId <目标部门ID>");
            return;
        }

        MergeDepartmentRequest request = new MergeDepartmentRequest();
        request.setTargetDepartmentId(params.get("targetId"));

        String path = "/api/departments/" + sourceId + "/merge";
        TypeReference<ApiResponse<Void>> typeRef = new TypeReference<ApiResponse<Void>>() {};
        ApiResponse<Void> response = httpClient.post(path, request, typeRef.getType());
        formatter.printResult(response);
    }

    private void getDepartmentTree(HttpClient httpClient, OutputFormatter formatter) throws Exception {
        String path = "/api/departments/tree";
        TypeReference<ApiResponse<List<DepartmentDTO>>> typeRef = new TypeReference<ApiResponse<List<DepartmentDTO>>>() {};
        ApiResponse<List<DepartmentDTO>> response = httpClient.get(path, typeRef.getType());

        if (response.getCode() == 0) {
            List<DepartmentDTO> tree = response.getData();
            if (tree == null || tree.isEmpty()) {
                System.out.println("暂无部门数据");
                return;
            }

            System.out.println("部门树:");
            System.out.println();
            printTree(tree, 0);
        } else {
            formatter.printResult(response);
        }
    }

    private void printTree(List<DepartmentDTO> departments, int level) {
        for (DepartmentDTO dept : departments) {
            String prefix = "  ".repeat(level);
            String connector = level > 0 ? "└─ " : "";
            System.out.println(prefix + connector + dept.getName() + " (" + dept.getId() + ")");
            
            if (dept.getChildren() != null && !dept.getChildren().isEmpty()) {
                printTree(dept.getChildren(), level + 1);
            }
        }
    }

    private void getDepartmentEmployees(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length == 0) {
            System.out.println("用法: department employees <部门ID> [--includeSub]");
            return;
        }

        String departmentId = args[0];
        boolean includeSub = Arrays.asList(args).contains("--includeSub");

        String path = "/api/employees/department/" + departmentId + "?includeSubDepartments=" + includeSub;
        TypeReference<ApiResponse<List<EmployeeDTO>>> typeRef = new TypeReference<ApiResponse<List<EmployeeDTO>>>() {};
        ApiResponse<List<EmployeeDTO>> response = httpClient.get(path, typeRef.getType());

        if (response.getCode() == 0) {
            List<EmployeeDTO> employees = response.getData();
            if (employees == null || employees.isEmpty()) {
                System.out.println("该部门暂无员工");
                return;
            }

            System.out.println("部门员工列表" + (includeSub ? " (包含子部门)" : "") + ":");
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
        System.out.println("部门管理命令:");
        System.out.println();
        System.out.println("  department list                    列出所有部门");
        System.out.println("  department get <id>               获取部门详情");
        System.out.println("  department create --name <名称> [--parentId <父部门ID>]");
        System.out.println("                                     创建部门");
        System.out.println("  department update <id> [--name <新名称>] [--parentId <新父部门ID>]");
        System.out.println("                                     更新部门");
        System.out.println("  department delete <id>             删除部门");
        System.out.println("  department merge <源部门ID> --targetId <目标部门ID>");
        System.out.println("                                     合并部门");
        System.out.println("  department tree                    显示部门树");
        System.out.println("  department employees <部门ID> [--includeSub]");
        System.out.println("                                     查询部门员工");
        System.out.println();
    }
}
