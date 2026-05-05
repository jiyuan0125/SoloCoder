package com.orgchart.client.command;

import com.fasterxml.jackson.core.type.TypeReference;
import com.orgchart.common.dto.ApiResponse;
import com.orgchart.common.dto.OperationLogDTO;
import com.orgchart.client.http.HttpClient;
import com.orgchart.client.util.OutputFormatter;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

public class LogCommands implements Command {

    @Override
    public String getName() {
        return "log";
    }

    @Override
    public String getDescription() {
        return "操作日志查询命令 (list, get)";
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
                listLogs(httpClient, formatter);
                break;
            case "get":
                getLogsByTarget(remainingArgs, httpClient, formatter);
                break;
            default:
                System.out.println("未知的子命令: " + subCommand);
                printHelp();
        }
    }

    private void listLogs(HttpClient httpClient, OutputFormatter formatter) throws Exception {
        String path = "/api/operation-logs";
        TypeReference<ApiResponse<List<OperationLogDTO>>> typeRef = new TypeReference<ApiResponse<List<OperationLogDTO>>>() {};
        ApiResponse<List<OperationLogDTO>> response = httpClient.get(path, typeRef.getType());

        if (response.getCode() == 0) {
            List<OperationLogDTO> logs = response.getData();
            if (logs == null || logs.isEmpty()) {
                System.out.println("暂无操作日志");
                return;
            }

            System.out.println("操作日志列表:");
            System.out.println();

            List<String> headers = Arrays.asList("时间", "操作类型", "目标类型", "目标名称", "变更");
            List<List<String>> rows = new ArrayList<>();
            for (OperationLogDTO log : logs) {
                List<String> row = new ArrayList<>();
                row.add(log.getOperationTime() != null ? log.getOperationTime().toString() : "-");
                row.add(log.getOperationType() != null ? log.getOperationType() : "-");
                row.add(log.getTargetType() != null ? log.getTargetType() : "-");
                row.add(log.getTargetName() != null ? log.getTargetName() : "-");
                String change = "";
                if (log.getOldValue() != null) {
                    change += "旧: " + log.getOldValue();
                }
                if (log.getNewValue() != null) {
                    if (!change.isEmpty()) {
                        change += " -> ";
                    }
                    change += "新: " + log.getNewValue();
                }
                row.add(change.isEmpty() ? "-" : change);
                rows.add(row);
            }
            formatter.printTable(headers, rows);
        } else {
            formatter.printResult(response);
        }
    }

    private void getLogsByTarget(String[] args, HttpClient httpClient, OutputFormatter formatter) throws Exception {
        if (args.length < 2) {
            System.out.println("用法: log get <目标类型> <目标ID>");
            System.out.println("  目标类型: DEPARTMENT, EMPLOYEE, VIRTUAL_TEAM");
            return;
        }

        String targetType = args[0];
        String targetId = args[1];
        String path = "/api/operation-logs/" + targetType + "/" + targetId;
        TypeReference<ApiResponse<List<OperationLogDTO>>> typeRef = new TypeReference<ApiResponse<List<OperationLogDTO>>>() {};
        ApiResponse<List<OperationLogDTO>> response = httpClient.get(path, typeRef.getType());

        if (response.getCode() == 0) {
            List<OperationLogDTO> logs = response.getData();
            if (logs == null || logs.isEmpty()) {
                System.out.println("该目标暂无操作日志");
                return;
            }

            System.out.println("目标操作日志:");
            System.out.println();

            List<String> headers = Arrays.asList("时间", "操作类型", "变更");
            List<List<String>> rows = new ArrayList<>();
            for (OperationLogDTO log : logs) {
                List<String> row = new ArrayList<>();
                row.add(log.getOperationTime() != null ? log.getOperationTime().toString() : "-");
                row.add(log.getOperationType() != null ? log.getOperationType() : "-");
                String change = "";
                if (log.getOldValue() != null) {
                    change += "旧: " + log.getOldValue();
                }
                if (log.getNewValue() != null) {
                    if (!change.isEmpty()) {
                        change += " -> ";
                    }
                    change += "新: " + log.getNewValue();
                }
                row.add(change.isEmpty() ? "-" : change);
                rows.add(row);
            }
            formatter.printTable(headers, rows);
        } else {
            formatter.printResult(response);
        }
    }

    private void printHelp() {
        System.out.println();
        System.out.println("操作日志命令:");
        System.out.println();
        System.out.println("  log list                         列出所有操作日志");
        System.out.println("  log get <目标类型> <目标ID>     查询目标的操作日志");
        System.out.println("                                   目标类型: DEPARTMENT, EMPLOYEE, VIRTUAL_TEAM");
        System.out.println();
    }
}
