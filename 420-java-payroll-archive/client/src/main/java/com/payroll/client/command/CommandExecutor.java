package com.payroll.client.command;

import com.fasterxml.jackson.core.type.TypeReference;
import com.payroll.client.http.HttpClient;
import com.payroll.client.util.OutputFormatter;
import com.payroll.common.dto.*;
import com.payroll.common.response.ApiResponse;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class CommandExecutor {

    private final HttpClient httpClient;

    public CommandExecutor() {
        this.httpClient = new HttpClient();
    }

    public void execute(String[] args) {
        if (args.length == 0 || "help".equals(args[0])) {
            OutputFormatter.printHelp();
            return;
        }

        String command = args[0];
        try {
            switch (command) {
                case "create":
                    executeCreate(args);
                    break;
                case "confirm":
                    executeConfirm(args);
                    break;
                case "remark":
                    executeRemark(args);
                    break;
                case "get":
                    executeGet(args);
                    break;
                case "query":
                    executeQuery(args);
                    break;
                case "report":
                    executeReport(args);
                    break;
                case "compare":
                    executeCompare(args);
                    break;
                default:
                    System.out.println("未知命令: " + command);
                    OutputFormatter.printHelp();
            }
        } catch (Exception e) {
            System.err.println("执行命令时发生错误: " + e.getMessage());
            e.printStackTrace();
        }
    }

    private void executeCreate(String[] args) throws IOException {
        if (args.length < 2) {
            System.out.println("用法: create <json>");
            System.out.println("示例: create '{\"employeeId\":\"E001\",\"employeeName\":\"张三\",\"idCard\":\"110101199001011234\",\"year\":2024,\"month\":1,\"baseSalary\":15000,\"deductionDetails\":[{\"deductionType\":\"社保\",\"amount\":1500},{\"deductionType\":\"个税\",\"amount\":300}]}'");
            return;
        }

        StringBuilder jsonBuilder = new StringBuilder();
        for (int i = 1; i < args.length; i++) {
            jsonBuilder.append(args[i]);
            if (i < args.length - 1) {
                jsonBuilder.append(" ");
            }
        }

        String json = jsonBuilder.toString().trim();
        if (json.startsWith("'") && json.endsWith("'")) {
            json = json.substring(1, json.length() - 1);
        }

        ArchiveRequestDTO request = httpClient.getObjectMapper().readValue(json, ArchiveRequestDTO.class);
        String response = httpClient.post("/api/archives", request);
        ApiResponse<PayrollArchiveDTO> apiResponse = httpClient.getObjectMapper().readValue(
                response,
                new TypeReference<ApiResponse<PayrollArchiveDTO>>() {}
        );

        if (apiResponse.getCode() == 0) {
            OutputFormatter.printSuccess("归档创建成功");
            if (apiResponse.getData() != null) {
                OutputFormatter.printArchive(apiResponse.getData());
            }
        } else {
            OutputFormatter.printError(apiResponse.getCode(), apiResponse.getMessage());
        }
    }

    private void executeConfirm(String[] args) throws IOException {
        if (args.length < 2) {
            System.out.println("用法: confirm <archiveId>");
            return;
        }

        String archiveId = args[1];
        String response = httpClient.post("/api/archives/" + archiveId + "/confirm", null);
        ApiResponse<PayrollArchiveDTO> apiResponse = httpClient.getObjectMapper().readValue(
                response,
                new TypeReference<ApiResponse<PayrollArchiveDTO>>() {}
        );

        if (apiResponse.getCode() == 0) {
            OutputFormatter.printSuccess("归档确认成功");
            if (apiResponse.getData() != null) {
                OutputFormatter.printArchive(apiResponse.getData());
            }
        } else {
            OutputFormatter.printError(apiResponse.getCode(), apiResponse.getMessage());
        }
    }

    private void executeRemark(String[] args) throws IOException {
        if (args.length < 3) {
            System.out.println("用法: remark <archiveId> <text>");
            return;
        }

        String archiveId = args[1];
        StringBuilder remarkBuilder = new StringBuilder();
        for (int i = 2; i < args.length; i++) {
            remarkBuilder.append(args[i]);
            if (i < args.length - 1) {
                remarkBuilder.append(" ");
            }
        }

        RemarkRequestDTO request = new RemarkRequestDTO(remarkBuilder.toString());
        String response = httpClient.post("/api/archives/" + archiveId + "/remarks", request);
        ApiResponse<PayrollArchiveDTO> apiResponse = httpClient.getObjectMapper().readValue(
                response,
                new TypeReference<ApiResponse<PayrollArchiveDTO>>() {}
        );

        if (apiResponse.getCode() == 0) {
            OutputFormatter.printSuccess("备注添加成功");
            if (apiResponse.getData() != null) {
                OutputFormatter.printArchive(apiResponse.getData());
            }
        } else {
            OutputFormatter.printError(apiResponse.getCode(), apiResponse.getMessage());
        }
    }

    private void executeGet(String[] args) throws IOException {
        if (args.length < 2) {
            System.out.println("用法: get <archiveId>");
            return;
        }

        String archiveId = args[1];
        String response = httpClient.get("/api/archives/" + archiveId, null);
        ApiResponse<PayrollArchiveDTO> apiResponse = httpClient.getObjectMapper().readValue(
                response,
                new TypeReference<ApiResponse<PayrollArchiveDTO>>() {}
        );

        if (apiResponse.getCode() == 0) {
            if (apiResponse.getData() != null) {
                OutputFormatter.printArchive(apiResponse.getData());
            } else {
                System.out.println("未找到记录");
            }
        } else {
            OutputFormatter.printError(apiResponse.getCode(), apiResponse.getMessage());
        }
    }

    private void executeQuery(String[] args) throws IOException {
        Map<String, String> params = new HashMap<>();
        
        for (int i = 1; i < args.length; i++) {
            String arg = args[i];
            if (arg.startsWith("--employee=")) {
                params.put("employeeId", arg.substring("--employee=".length()));
            } else if (arg.startsWith("--year=")) {
                params.put("year", arg.substring("--year=".length()));
            } else if (arg.startsWith("--month=")) {
                params.put("month", arg.substring("--month=".length()));
            }
        }

        String response = httpClient.get("/api/archives", params);
        ApiResponse<List<PayrollArchiveDTO>> apiResponse = httpClient.getObjectMapper().readValue(
                response,
                new TypeReference<ApiResponse<List<PayrollArchiveDTO>>>() {}
        );

        if (apiResponse.getCode() == 0) {
            OutputFormatter.printArchiveList(apiResponse.getData());
        } else {
            OutputFormatter.printError(apiResponse.getCode(), apiResponse.getMessage());
        }
    }

    private void executeReport(String[] args) throws IOException {
        if (args.length < 2) {
            System.out.println("用法: report <year>");
            return;
        }

        int year;
        try {
            year = Integer.parseInt(args[1]);
        } catch (NumberFormatException e) {
            System.out.println("年份格式无效");
            return;
        }

        String response = httpClient.get("/api/archives/years/" + year + "/report", null);
        ApiResponse<AnnualReportDTO> apiResponse = httpClient.getObjectMapper().readValue(
                response,
                new TypeReference<ApiResponse<AnnualReportDTO>>() {}
        );

        if (apiResponse.getCode() == 0) {
            if (apiResponse.getData() != null) {
                OutputFormatter.printAnnualReport(apiResponse.getData());
            }
        } else {
            OutputFormatter.printError(apiResponse.getCode(), apiResponse.getMessage());
        }
    }

    private void executeCompare(String[] args) throws IOException {
        if (args.length < 3) {
            System.out.println("用法: compare <baseYear> <compareYear>");
            return;
        }

        int baseYear, compareYear;
        try {
            baseYear = Integer.parseInt(args[1]);
            compareYear = Integer.parseInt(args[2]);
        } catch (NumberFormatException e) {
            System.out.println("年份格式无效");
            return;
        }

        Map<String, String> params = new HashMap<>();
        params.put("baseYear", String.valueOf(baseYear));
        params.put("compareYear", String.valueOf(compareYear));

        String response = httpClient.get("/api/archives/compare", params);
        ApiResponse<YearComparisonDTO> apiResponse = httpClient.getObjectMapper().readValue(
                response,
                new TypeReference<ApiResponse<YearComparisonDTO>>() {}
        );

        if (apiResponse.getCode() == 0) {
            if (apiResponse.getData() != null) {
                OutputFormatter.printYearComparison(apiResponse.getData());
            }
        } else {
            OutputFormatter.printError(apiResponse.getCode(), apiResponse.getMessage());
        }
    }
}
