package com.performance.client;

import com.fasterxml.jackson.core.type.TypeReference;
import com.performance.client.util.HttpClientUtil;
import com.performance.common.dto.ApiResponse;
import com.performance.common.dto.DepartmentDTO;
import com.performance.common.dto.EmployeeDTO;
import com.performance.common.dto.PerformanceReviewDTO;
import com.performance.common.dto.ReviewCycleDTO;
import com.performance.common.request.ConfirmReviewRequest;
import com.performance.common.request.CreateCycleRequest;
import com.performance.common.request.QueryHistoryRequest;
import com.performance.common.request.SubmitManagerReviewRequest;
import com.performance.common.request.SubmitSelfReviewRequest;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.math.BigDecimal;
import java.time.LocalDate;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class PerformanceReviewClient {
    private static final BufferedReader READER = new BufferedReader(new InputStreamReader(System.in));

    public static void main(String[] args) {
        System.out.println("=========================================");
        System.out.println("  绩效评估系统客户端");
        System.out.println("=========================================");
        
        try {
            showMainMenu();
        } catch (Exception e) {
            System.err.println("发生错误: " + e.getMessage());
            e.printStackTrace();
        }
    }

    private static void showMainMenu() throws Exception {
        while (true) {
            System.out.println("\n--- 主菜单 ---");
            System.out.println("1. 部门管理");
            System.out.println("2. 员工管理");
            System.out.println("3. 评估周期管理 (HR)");
            System.out.println("4. 员工自评");
            System.out.println("5. 上级评分");
            System.out.println("6. HR确认评估");
            System.out.println("7. 查看我的绩效");
            System.out.println("8. 历史评估查询");
            System.out.println("0. 退出");
            System.out.print("请选择操作: ");
            
            String choice = READER.readLine();
            switch (choice) {
                case "1":
                    departmentMenu();
                    break;
                case "2":
                    employeeMenu();
                    break;
                case "3":
                    cycleMenu();
                    break;
                case "4":
                    submitSelfReview();
                    break;
                case "5":
                    submitManagerReview();
                    break;
                case "6":
                    confirmReview();
                    break;
                case "7":
                    viewMyPerformance();
                    break;
                case "8":
                    queryHistory();
                    break;
                case "0":
                    System.out.println("再见!");
                    return;
                default:
                    System.out.println("无效选择，请重新输入");
            }
        }
    }

    private static void departmentMenu() throws Exception {
        while (true) {
            System.out.println("\n--- 部门管理 ---");
            System.out.println("1. 创建部门");
            System.out.println("2. 查看所有部门");
            System.out.println("3. 查看部门详情");
            System.out.println("0. 返回");
            System.out.print("请选择操作: ");
            
            String choice = READER.readLine();
            switch (choice) {
                case "1":
                    createDepartment();
                    break;
                case "2":
                    listAllDepartments();
                    break;
                case "3":
                    getDepartmentById();
                    break;
                case "0":
                    return;
                default:
                    System.out.println("无效选择");
            }
        }
    }

    private static void createDepartment() throws Exception {
        System.out.print("输入部门名称: ");
        String name = READER.readLine();
        System.out.print("输入部门编码: ");
        String code = READER.readLine();
        
        DepartmentDTO dept = new DepartmentDTO();
        dept.setName(name);
        dept.setCode(code);
        
        ApiResponse<DepartmentDTO> response = HttpClientUtil.post(
                "/api/departments", 
                dept, 
                new TypeReference<ApiResponse<DepartmentDTO>>() {});
        
        if (response.getCode() == 0) {
            System.out.println("部门创建成功!");
            System.out.println(HttpClientUtil.toJson(response.getData()));
        } else {
            System.out.println("创建失败: " + response.getMessage());
        }
    }

    private static void listAllDepartments() throws Exception {
        ApiResponse<List<DepartmentDTO>> response = HttpClientUtil.get(
                "/api/departments", 
                new TypeReference<ApiResponse<List<DepartmentDTO>>>() {});
        
        if (response.getCode() == 0 && response.getData() != null) {
            System.out.println("\n部门列表:");
            for (DepartmentDTO dept : response.getData()) {
                System.out.printf("ID: %d, 名称: %s, 编码: %s%n", 
                        dept.getId(), dept.getName(), dept.getCode());
            }
        } else {
            System.out.println("查询失败: " + response.getMessage());
        }
    }

    private static void getDepartmentById() throws Exception {
        System.out.print("输入部门ID: ");
        Long id = Long.parseLong(READER.readLine());
        
        ApiResponse<DepartmentDTO> response = HttpClientUtil.get(
                "/api/departments/" + id, 
                new TypeReference<ApiResponse<DepartmentDTO>>() {});
        
        if (response.getCode() == 0 && response.getData() != null) {
            System.out.println(HttpClientUtil.toJson(response.getData()));
        } else {
            System.out.println("查询失败: " + response.getMessage());
        }
    }

    private static void employeeMenu() throws Exception {
        while (true) {
            System.out.println("\n--- 员工管理 ---");
            System.out.println("1. 创建员工");
            System.out.println("2. 查看所有员工");
            System.out.println("3. 查看员工详情");
            System.out.println("4. 员工调岗");
            System.out.println("0. 返回");
            System.out.print("请选择操作: ");
            
            String choice = READER.readLine();
            switch (choice) {
                case "1":
                    createEmployee();
                    break;
                case "2":
                    listAllEmployees();
                    break;
                case "3":
                    getEmployeeById();
                    break;
                case "4":
                    transferEmployee();
                    break;
                case "0":
                    return;
                default:
                    System.out.println("无效选择");
            }
        }
    }

    private static void createEmployee() throws Exception {
        System.out.print("输入员工姓名: ");
        String name = READER.readLine();
        System.out.print("输入员工工号: ");
        String employeeNo = READER.readLine();
        System.out.print("输入部门ID: ");
        Long deptId = Long.parseLong(READER.readLine());
        System.out.print("输入上级ID (可选，直接回车跳过): ");
        String managerIdStr = READER.readLine();
        Long managerId = managerIdStr.isEmpty() ? null : Long.parseLong(managerIdStr);
        System.out.print("输入角色: ");
        String role = READER.readLine();
        
        EmployeeDTO employee = new EmployeeDTO();
        employee.setName(name);
        employee.setEmployeeNo(employeeNo);
        employee.setDepartmentId(deptId);
        employee.setManagerId(managerId);
        employee.setRole(role);
        
        ApiResponse<EmployeeDTO> response = HttpClientUtil.post(
                "/api/employees", 
                employee, 
                new TypeReference<ApiResponse<EmployeeDTO>>() {});
        
        if (response.getCode() == 0) {
            System.out.println("员工创建成功!");
            System.out.println(HttpClientUtil.toJson(response.getData()));
        } else {
            System.out.println("创建失败: " + response.getMessage());
        }
    }

    private static void listAllEmployees() throws Exception {
        ApiResponse<List<EmployeeDTO>> response = HttpClientUtil.get(
                "/api/employees", 
                new TypeReference<ApiResponse<List<EmployeeDTO>>>() {});
        
        if (response.getCode() == 0 && response.getData() != null) {
            System.out.println("\n员工列表:");
            for (EmployeeDTO emp : response.getData()) {
                System.out.printf("ID: %d, 姓名: %s, 工号: %s, 部门ID: %d, 角色: %s%n", 
                        emp.getId(), emp.getName(), emp.getEmployeeNo(), 
                        emp.getDepartmentId(), emp.getRole());
            }
        } else {
            System.out.println("查询失败: " + response.getMessage());
        }
    }

    private static void getEmployeeById() throws Exception {
        System.out.print("输入员工ID: ");
        Long id = Long.parseLong(READER.readLine());
        
        ApiResponse<EmployeeDTO> response = HttpClientUtil.get(
                "/api/employees/" + id, 
                new TypeReference<ApiResponse<EmployeeDTO>>() {});
        
        if (response.getCode() == 0 && response.getData() != null) {
            System.out.println(HttpClientUtil.toJson(response.getData()));
        } else {
            System.out.println("查询失败: " + response.getMessage());
        }
    }

    private static void transferEmployee() throws Exception {
        System.out.print("输入员工ID: ");
        Long employeeId = Long.parseLong(READER.readLine());
        System.out.print("输入新部门ID: ");
        Long newDeptId = Long.parseLong(READER.readLine());
        System.out.print("输入新上级ID: ");
        Long newManagerId = Long.parseLong(READER.readLine());
        
        Map<String, Object> params = new HashMap<>();
        
        ApiResponse<EmployeeDTO> response = HttpClientUtil.put(
                "/api/employees/" + employeeId + "/transfer?newDepartmentId=" + newDeptId + "&newManagerId=" + newManagerId,
                null,
                new TypeReference<ApiResponse<EmployeeDTO>>() {});
        
        if (response.getCode() == 0) {
            System.out.println("调岗成功!");
            System.out.println(HttpClientUtil.toJson(response.getData()));
        } else {
            System.out.println("调岗失败: " + response.getMessage());
        }
    }

    private static void cycleMenu() throws Exception {
        while (true) {
            System.out.println("\n--- 评估周期管理 (HR) ---");
            System.out.println("1. 创建评估周期");
            System.out.println("2. 查看所有周期");
            System.out.println("3. 推进到下一阶段");
            System.out.println("0. 返回");
            System.out.print("请选择操作: ");
            
            String choice = READER.readLine();
            switch (choice) {
                case "1":
                    createCycle();
                    break;
                case "2":
                    listAllCycles();
                    break;
                case "3":
                    advanceCycle();
                    break;
                case "0":
                    return;
                default:
                    System.out.println("无效选择");
            }
        }
    }

    private static void createCycle() throws Exception {
        System.out.print("输入年份: ");
        String year = READER.readLine();
        System.out.print("输入季度 (1/2/3/4): ");
        String quarter = READER.readLine();
        
        CreateCycleRequest request = new CreateCycleRequest();
        request.setYear(year);
        request.setQuarter(quarter);
        request.setStartDate(LocalDate.now());
        request.setEndDate(LocalDate.now().plusMonths(3));
        
        ApiResponse<Void> response = HttpClientUtil.post(
                "/api/cycles", 
                request, 
                new TypeReference<ApiResponse<Void>>() {});
        
        if (response.getCode() == 0) {
            System.out.println("评估周期创建成功!");
        } else {
            System.out.println("创建失败: " + response.getMessage());
        }
    }

    private static void listAllCycles() throws Exception {
        ApiResponse<List<ReviewCycleDTO>> response = HttpClientUtil.get(
                "/api/cycles", 
                new TypeReference<ApiResponse<List<ReviewCycleDTO>>>() {});
        
        if (response.getCode() == 0 && response.getData() != null) {
            System.out.println("\n评估周期列表:");
            for (ReviewCycleDTO cycle : response.getData()) {
                System.out.printf("ID: %d, 名称: %s, 状态: %s%n", 
                        cycle.getId(), cycle.getCycleName(), cycle.getStatus());
            }
        } else {
            System.out.println("查询失败: " + response.getMessage());
        }
    }

    private static void advanceCycle() throws Exception {
        System.out.print("输入周期ID: ");
        Long cycleId = Long.parseLong(READER.readLine());
        
        ApiResponse<Void> response = HttpClientUtil.put(
                "/api/cycles/" + cycleId + "/advance",
                null,
                new TypeReference<ApiResponse<Void>>() {});
        
        if (response.getCode() == 0) {
            System.out.println("已推进到下一阶段!");
        } else {
            System.out.println("操作失败: " + response.getMessage());
        }
    }

    private static void submitSelfReview() throws Exception {
        System.out.print("输入周期ID: ");
        Long cycleId = Long.parseLong(READER.readLine());
        System.out.print("输入员工ID: ");
        Long employeeId = Long.parseLong(READER.readLine());
        System.out.print("输入自评分数 (1-5): ");
        BigDecimal score = new BigDecimal(READER.readLine());
        System.out.print("输入自评评语: ");
        String comment = READER.readLine();
        
        SubmitSelfReviewRequest request = new SubmitSelfReviewRequest();
        request.setCycleId(cycleId);
        request.setEmployeeId(employeeId);
        request.setScore(score);
        request.setComment(comment);
        
        ApiResponse<Void> response = HttpClientUtil.post(
                "/api/reviews/self-review", 
                request, 
                new TypeReference<ApiResponse<Void>>() {});
        
        if (response.getCode() == 0) {
            System.out.println("自评提交成功!");
        } else {
            System.out.println("提交失败: " + response.getMessage());
        }
    }

    private static void submitManagerReview() throws Exception {
        System.out.print("输入周期ID: ");
        Long cycleId = Long.parseLong(READER.readLine());
        System.out.print("输入上级ID: ");
        Long managerId = Long.parseLong(READER.readLine());
        System.out.print("输入员工ID: ");
        Long employeeId = Long.parseLong(READER.readLine());
        System.out.print("输入评分 (1-5，必须与自评分至少差0.5): ");
        BigDecimal score = new BigDecimal(READER.readLine());
        System.out.print("输入评语: ");
        String comment = READER.readLine();
        
        SubmitManagerReviewRequest request = new SubmitManagerReviewRequest();
        request.setCycleId(cycleId);
        request.setManagerId(managerId);
        request.setEmployeeId(employeeId);
        request.setScore(score);
        request.setComment(comment);
        
        ApiResponse<Void> response = HttpClientUtil.post(
                "/api/reviews/manager-review", 
                request, 
                new TypeReference<ApiResponse<Void>>() {});
        
        if (response.getCode() == 0) {
            System.out.println("评分提交成功!");
        } else {
            System.out.println("提交失败: " + response.getMessage());
        }
    }

    private static void confirmReview() throws Exception {
        System.out.print("输入周期ID: ");
        Long cycleId = Long.parseLong(READER.readLine());
        System.out.print("输入员工ID: ");
        Long employeeId = Long.parseLong(READER.readLine());
        
        ConfirmReviewRequest request = new ConfirmReviewRequest();
        request.setCycleId(cycleId);
        request.setEmployeeId(employeeId);
        
        ApiResponse<Void> response = HttpClientUtil.post(
                "/api/reviews/confirm", 
                request, 
                new TypeReference<ApiResponse<Void>>() {});
        
        if (response.getCode() == 0) {
            System.out.println("确认成功!");
        } else {
            System.out.println("确认失败: " + response.getMessage());
        }
    }

    private static void viewMyPerformance() throws Exception {
        System.out.print("输入员工ID: ");
        Long employeeId = Long.parseLong(READER.readLine());
        
        ApiResponse<List<PerformanceReviewDTO>> response = HttpClientUtil.get(
                "/api/reviews/employee/" + employeeId, 
                new TypeReference<ApiResponse<List<PerformanceReviewDTO>>>() {});
        
        if (response.getCode() == 0 && response.getData() != null) {
            System.out.println("\n我的绩效记录:");
            for (PerformanceReviewDTO review : response.getData()) {
                System.out.println("-----------------------------------");
                System.out.printf("周期: %s%n", review.getCycleName());
                System.out.printf("最终得分: %s%n", review.getFinalScore());
                System.out.printf("等级: %s%n", review.getFinalGrade() != null ? review.getFinalGrade().getGrade() : "N/A");
                System.out.printf("年终奖系数: %s%n", review.getBonusCoefficient());
                if (review.getPercentile() != null) {
                    System.out.printf("部门排名: 第%d名 / 共%d人%n", review.getRanking(), review.getTotalEmployees());
                    System.out.printf("百分位: %s%%%n", review.getPercentile());
                }
                if (review.getPipEligible() != null && review.getPipEligible()) {
                    System.out.println("⚠️  注意: 已进入绩效改进计划 (PIP)");
                }
                if (review.getTerminationEligible() != null && review.getTerminationEligible()) {
                    System.out.println("⚠️  警告: 已触发辞退流程");
                }
            }
        } else {
            System.out.println("查询失败: " + response.getMessage());
        }
    }

    private static void queryHistory() throws Exception {
        QueryHistoryRequest request = new QueryHistoryRequest();
        
        System.out.print("输入部门ID (可选，直接回车跳过): ");
        String deptIdStr = READER.readLine();
        if (!deptIdStr.isEmpty()) {
            request.setDepartmentId(Long.parseLong(deptIdStr));
        }
        
        System.out.print("输入年份 (可选，直接回车跳过): ");
        String year = READER.readLine();
        if (!year.isEmpty()) {
            request.setYear(year);
        }
        
        System.out.print("输入季度 (可选，直接回车跳过): ");
        String quarter = READER.readLine();
        if (!quarter.isEmpty()) {
            request.setQuarter(quarter);
        }
        
        System.out.print("输入员工ID (可选，直接回车跳过): ");
        String empIdStr = READER.readLine();
        if (!empIdStr.isEmpty()) {
            request.setEmployeeId(Long.parseLong(empIdStr));
        }
        
        ApiResponse<List<PerformanceReviewDTO>> response = HttpClientUtil.post(
                "/api/reviews/history", 
                request, 
                new TypeReference<ApiResponse<List<PerformanceReviewDTO>>>() {});
        
        if (response.getCode() == 0 && response.getData() != null) {
            System.out.println("\n历史评估记录:");
            for (PerformanceReviewDTO review : response.getData()) {
                System.out.println("-----------------------------------");
                System.out.printf("员工: %s%n", review.getEmployeeName());
                System.out.printf("周期: %s%n", review.getCycleName());
                System.out.printf("最终得分: %s%n", review.getFinalScore());
                System.out.printf("等级: %s%n", review.getFinalGrade() != null ? review.getFinalGrade().getGrade() : "N/A");
                System.out.printf("年终奖系数: %s%n", review.getBonusCoefficient());
            }
        } else {
            System.out.println("查询失败: " + response.getMessage());
        }
    }
}