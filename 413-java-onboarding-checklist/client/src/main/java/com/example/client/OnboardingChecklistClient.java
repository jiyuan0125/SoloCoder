package com.example.client;

import com.example.client.command.CommandParser;
import com.example.client.http.HttpClient;
import com.example.client.output.OutputFormatter;
import com.example.common.dto.*;
import com.example.common.enums.PositionType;
import com.example.common.request.*;
import com.example.common.response.ApiResponse;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import java.time.LocalDate;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class OnboardingChecklistClient {
    private static final String DEFAULT_BASE_URL = "http://localhost:8080";
    
    private final HttpClient httpClient;
    private final CommandParser commandParser;
    private final OutputFormatter outputFormatter;
    private final ObjectMapper objectMapper;

    public OnboardingChecklistClient(String baseUrl) {
        this.httpClient = new HttpClient(baseUrl);
        this.commandParser = new CommandParser();
        this.outputFormatter = new OutputFormatter();
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
    }

    public void run(String[] args) {
        CommandParser.Command command = commandParser.parse(args);
        String action = command.getAction();

        try {
            switch (action) {
                case "help":
                case "-h":
                case "--help":
                    outputFormatter.printHelp();
                    break;

                case "employee-list":
                    handleEmployeeList();
                    break;

                case "employee-get":
                    handleEmployeeGet(command);
                    break;

                case "employee-create":
                    handleEmployeeCreate(command);
                    break;

                case "employee-delete":
                    handleEmployeeDelete(command);
                    break;

                case "item-add":
                    handleItemAdd(command);
                    break;

                case "item-update":
                    handleItemUpdate(command);
                    break;

                case "template-list":
                    handleTemplateList();
                    break;

                case "template-get":
                    handleTemplateGet(command);
                    break;

                case "template-copy":
                    handleTemplateCopy(command);
                    break;

                case "alert-list":
                    handleAlertList();
                    break;

                case "alert-unread":
                    handleAlertUnread();
                    break;

                case "alert-get":
                    handleAlertGet(command);
                    break;

                case "alert-read":
                    handleAlertRead(command);
                    break;

                case "stats":
                    handleStats();
                    break;

                default:
                    outputFormatter.printError("未知命令: " + action);
                    outputFormatter.printHelp();
            }
        } catch (Exception e) {
            outputFormatter.printError("执行出错: " + e.getMessage());
            e.printStackTrace();
        }
    }

    private void handleEmployeeList() throws Exception {
        String response = httpClient.get("/api/employees", String.class);
        ApiResponse<List<EmployeeDTO>> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<List<EmployeeDTO>>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printEmployeeList(apiResponse.getData());
        } else {
            outputFormatter.printError("获取员工列表失败: " + apiResponse.getMessage());
        }
    }

    private void handleEmployeeGet(CommandParser.Command command) throws Exception {
        String id = command.getParameter("id");
        if (id == null) {
            outputFormatter.printError("缺少参数: --id");
            return;
        }

        String response = httpClient.get("/api/employees/" + id, String.class);
        ApiResponse<EmployeeDTO> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<EmployeeDTO>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printEmployee(apiResponse.getData());
        } else {
            outputFormatter.printError("获取员工详情失败: " + apiResponse.getMessage());
        }
    }

    private void handleEmployeeCreate(CommandParser.Command command) throws Exception {
        String name = command.getParameter("name");
        String positionStr = command.getParameter("position");
        LocalDate onboardingDate = command.getParameterAsDate("onboarding-date");

        if (name == null) {
            outputFormatter.printError("缺少参数: --name");
            return;
        }
        if (positionStr == null) {
            outputFormatter.printError("缺少参数: --position");
            return;
        }
        if (onboardingDate == null) {
            outputFormatter.printError("缺少或格式错误: --onboarding-date (格式: YYYY-MM-DD)");
            return;
        }

        PositionType positionType;
        try {
            positionType = PositionType.valueOf(positionStr.toUpperCase());
        } catch (IllegalArgumentException e) {
            outputFormatter.printError("无效的岗位类型: " + positionStr);
            return;
        }

        CreateEmployeeRequest request = new CreateEmployeeRequest();
        request.setName(name);
        request.setEmail(command.getParameter("email"));
        request.setPositionType(positionType);
        request.setDepartment(command.getParameter("department"));
        request.setOnboardingDate(onboardingDate);

        String response = httpClient.post("/api/employees", request, String.class);
        ApiResponse<EmployeeDTO> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<EmployeeDTO>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printSuccess("员工创建成功");
            outputFormatter.printEmployee(apiResponse.getData());
        } else {
            outputFormatter.printError("创建员工失败: " + apiResponse.getMessage());
        }
    }

    private void handleEmployeeDelete(CommandParser.Command command) throws Exception {
        String id = command.getParameter("id");
        if (id == null) {
            outputFormatter.printError("缺少参数: --id");
            return;
        }

        String response = httpClient.delete("/api/employees/" + id, String.class);
        ApiResponse<Void> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<Void>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printSuccess("员工删除成功");
        } else {
            outputFormatter.printError("删除员工失败: " + apiResponse.getMessage());
        }
    }

    private void handleItemAdd(CommandParser.Command command) throws Exception {
        String employeeId = command.getParameter("employee-id");
        String name = command.getParameter("name");
        String responsible = command.getParameter("responsible");
        LocalDate dueDate = command.getParameterAsDate("due-date");

        if (employeeId == null) {
            outputFormatter.printError("缺少参数: --employee-id");
            return;
        }
        if (name == null) {
            outputFormatter.printError("缺少参数: --name");
            return;
        }
        if (responsible == null) {
            outputFormatter.printError("缺少参数: --responsible");
            return;
        }
        if (dueDate == null) {
            outputFormatter.printError("缺少或格式错误: --due-date (格式: YYYY-MM-DD)");
            return;
        }

        CreateChecklistItemRequest request = new CreateChecklistItemRequest();
        request.setName(name);
        request.setDescription(command.getParameter("description"));
        request.setResponsiblePerson(responsible);
        request.setDepartment(command.getParameter("department"));
        request.setDueDate(dueDate);
        request.setRequired(command.getParameterAsBoolean("required"));
        request.setOnboardingDayRequired(command.getParameterAsBoolean("onboarding-day-required"));
        request.setPreOnboarding(command.getParameterAsBoolean("pre-onboarding"));

        String response = httpClient.post("/api/employees/" + employeeId + "/items", request, String.class);
        ApiResponse<ChecklistItemDTO> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<ChecklistItemDTO>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printSuccess("事项添加成功");
            outputFormatter.printChecklistItem(apiResponse.getData());
        } else {
            outputFormatter.printError("添加事项失败: " + apiResponse.getMessage());
        }
    }

    private void handleItemUpdate(CommandParser.Command command) throws Exception {
        String employeeId = command.getParameter("employee-id");
        String itemId = command.getParameter("item-id");

        if (employeeId == null) {
            outputFormatter.printError("缺少参数: --employee-id");
            return;
        }
        if (itemId == null) {
            outputFormatter.printError("缺少参数: --item-id");
            return;
        }

        UpdateChecklistItemRequest request = new UpdateChecklistItemRequest();
        if (command.hasParameter("name")) request.setName(command.getParameter("name"));
        if (command.hasParameter("description")) request.setDescription(command.getParameter("description"));
        if (command.hasParameter("responsible")) request.setResponsiblePerson(command.getParameter("responsible"));
        if (command.hasParameter("department")) request.setDepartment(command.getParameter("department"));
        if (command.hasParameter("due-date")) request.setDueDate(command.getParameterAsDate("due-date"));
        if (command.hasParameter("completed")) request.setIsCompleted(command.getParameterAsBoolean("completed"));

        String response = httpClient.put("/api/employees/" + employeeId + "/items/" + itemId, request, String.class);
        ApiResponse<ChecklistItemDTO> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<ChecklistItemDTO>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printSuccess("事项更新成功");
            outputFormatter.printChecklistItem(apiResponse.getData());
        } else {
            outputFormatter.printError("更新事项失败: " + apiResponse.getMessage());
        }
    }

    private void handleTemplateList() throws Exception {
        String response = httpClient.get("/api/templates", String.class);
        ApiResponse<List<ChecklistTemplateDTO>> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<List<ChecklistTemplateDTO>>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printTemplateList(apiResponse.getData());
        } else {
            outputFormatter.printError("获取模板列表失败: " + apiResponse.getMessage());
        }
    }

    private void handleTemplateGet(CommandParser.Command command) throws Exception {
        String id = command.getParameter("id");
        if (id == null) {
            outputFormatter.printError("缺少参数: --id");
            return;
        }

        String response = httpClient.get("/api/templates/" + id, String.class);
        ApiResponse<ChecklistTemplateDTO> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<ChecklistTemplateDTO>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printTemplate(apiResponse.getData());
        } else {
            outputFormatter.printError("获取模板详情失败: " + apiResponse.getMessage());
        }
    }

    private void handleTemplateCopy(CommandParser.Command command) throws Exception {
        String sourceId = command.getParameter("source-id");
        String newName = command.getParameter("new-name");

        if (sourceId == null) {
            outputFormatter.printError("缺少参数: --source-id");
            return;
        }
        if (newName == null) {
            outputFormatter.printError("缺少参数: --new-name");
            return;
        }

        String response = httpClient.post("/api/templates/" + sourceId + "/copy?newName=" + newName, null, String.class);
        ApiResponse<ChecklistTemplateDTO> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<ChecklistTemplateDTO>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printSuccess("模板复制成功");
            outputFormatter.printTemplate(apiResponse.getData());
        } else {
            outputFormatter.printError("复制模板失败: " + apiResponse.getMessage());
        }
    }

    private void handleAlertList() throws Exception {
        String response = httpClient.get("/api/alerts", String.class);
        ApiResponse<List<AlertDTO>> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<List<AlertDTO>>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printAlertList(apiResponse.getData());
        } else {
            outputFormatter.printError("获取告警列表失败: " + apiResponse.getMessage());
        }
    }

    private void handleAlertUnread() throws Exception {
        String response = httpClient.get("/api/alerts/unread", String.class);
        ApiResponse<List<AlertDTO>> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<List<AlertDTO>>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printAlertList(apiResponse.getData());
        } else {
            outputFormatter.printError("获取未读告警失败: " + apiResponse.getMessage());
        }
    }

    private void handleAlertGet(CommandParser.Command command) throws Exception {
        String id = command.getParameter("id");
        if (id == null) {
            outputFormatter.printError("缺少参数: --id");
            return;
        }

        String response = httpClient.get("/api/alerts/" + id, String.class);
        ApiResponse<AlertDTO> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<AlertDTO>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printAlert(apiResponse.getData());
        } else {
            outputFormatter.printError("获取告警详情失败: " + apiResponse.getMessage());
        }
    }

    private void handleAlertRead(CommandParser.Command command) throws Exception {
        String id = command.getParameter("id");
        if (id == null) {
            outputFormatter.printError("缺少参数: --id");
            return;
        }

        String response = httpClient.put("/api/alerts/" + id + "/read", null, String.class);
        ApiResponse<Void> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<Void>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printSuccess("告警已标记为已读");
        } else {
            outputFormatter.printError("标记告警失败: " + apiResponse.getMessage());
        }
    }

    private void handleStats() throws Exception {
        String response = httpClient.get("/api/statistics", String.class);
        ApiResponse<OnboardingStatisticsDTO> apiResponse = objectMapper.readValue(response,
                new TypeReference<ApiResponse<OnboardingStatisticsDTO>>() {});
        
        if (apiResponse.getCode() == 0) {
            outputFormatter.printStatistics(apiResponse.getData());
        } else {
            outputFormatter.printError("获取统计信息失败: " + apiResponse.getMessage());
        }
    }

    public static void main(String[] args) {
        String baseUrl = System.getenv("SERVER_URL");
        if (baseUrl == null || baseUrl.isEmpty()) {
            baseUrl = DEFAULT_BASE_URL;
        }

        OnboardingChecklistClient client = new OnboardingChecklistClient(baseUrl);
        client.run(args);
    }
}
