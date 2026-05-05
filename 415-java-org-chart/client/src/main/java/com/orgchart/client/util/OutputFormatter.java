package com.orgchart.client.util;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.orgchart.common.dto.ApiResponse;

import java.util.List;
import java.util.Map;

public class OutputFormatter {

    private final ObjectMapper objectMapper;

    public OutputFormatter(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    public void printResult(ApiResponse<?> response) {
        if (response.getCode() == 0) {
            System.out.println("操作成功!");
            if (response.getData() != null) {
                try {
                    String json = objectMapper.writerWithDefaultPrettyPrinter().writeValueAsString(response.getData());
                    System.out.println("返回数据:");
                    System.out.println(json);
                } catch (JsonProcessingException e) {
                    System.out.println("返回数据: " + response.getData());
                }
            }
        } else {
            System.out.println("操作失败!");
            System.out.println("错误码: " + response.getCode());
            System.out.println("错误信息: " + response.getMessage());
        }
    }

    public void printSuccess(String message) {
        System.out.println("✓ " + message);
    }

    public void printError(String message) {
        System.out.println("✗ " + message);
    }

    public void printTable(List<String> headers, List<List<String>> rows) {
        if (headers.isEmpty() && rows.isEmpty()) {
            return;
        }

        int[] colWidths = new int[headers.size()];
        for (int i = 0; i < headers.size(); i++) {
            colWidths[i] = headers.get(i).length();
        }

        for (List<String> row : rows) {
            for (int i = 0; i < row.size(); i++) {
                if (i < colWidths.length) {
                    int len = row.get(i) == null ? 0 : row.get(i).length();
                    colWidths[i] = Math.max(colWidths[i], len);
                }
            }
        }

        printSeparator(colWidths);
        printRow(headers, colWidths);
        printSeparator(colWidths);

        for (List<String> row : rows) {
            printRow(row, colWidths);
        }

        if (!rows.isEmpty()) {
            printSeparator(colWidths);
        }
    }

    private void printSeparator(int[] colWidths) {
        StringBuilder sb = new StringBuilder();
        sb.append("+");
        for (int width : colWidths) {
            sb.append("-".repeat(width + 2)).append("+");
        }
        System.out.println(sb);
    }

    private void printRow(List<String> row, int[] colWidths) {
        StringBuilder sb = new StringBuilder();
        sb.append("|");
        for (int i = 0; i < colWidths.length; i++) {
            String value = i < row.size() && row.get(i) != null ? row.get(i) : "";
            sb.append(" ").append(value).append(" ".repeat(colWidths[i] - value.length())).append(" |");
        }
        System.out.println(sb);
    }

    public void printJson(Object obj) {
        try {
            String json = objectMapper.writerWithDefaultPrettyPrinter().writeValueAsString(obj);
            System.out.println(json);
        } catch (JsonProcessingException e) {
            System.out.println(obj);
        }
    }

    public void printKeyValue(Map<String, String> map) {
        int maxKeyLen = map.keySet().stream().mapToInt(String::length).max().orElse(0);
        for (Map.Entry<String, String> entry : map.entrySet()) {
            String key = entry.getKey();
            String value = entry.getValue() == null ? "" : entry.getValue();
            System.out.println("  " + key + ": " + " ".repeat(maxKeyLen - key.length()) + value);
        }
    }
}
