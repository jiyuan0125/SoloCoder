package com.workshift.client.command;

import com.fasterxml.jackson.core.type.TypeReference;
import com.workshift.common.dto.ShiftDTO;
import com.workshift.common.enums.ShiftType;
import com.workshift.common.request.CreateShiftRequest;
import com.workshift.common.response.ApiResponse;
import com.workshift.client.http.HttpClient;
import com.workshift.client.util.OutputFormatter;
import java.time.LocalDate;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class ShiftCommand implements Command {

    private final HttpClient httpClient;

    public ShiftCommand(HttpClient httpClient) {
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
                if (args.length >= 3) {
                    listShiftsByEmployee(args[2]);
                } else {
                    listAllShifts();
                }
                break;
            case "range":
                if (args.length < 4) {
                    System.out.println("用法: shift range <startDate> <endDate>");
                    return;
                }
                listShiftsByDateRange(args[2], args[3]);
                break;
            case "create":
                if (args.length < 5) {
                    System.out.println("用法: shift create <employeeId> <date> <shiftType>");
                    System.out.println("班次类型: MORNING, AFTERNOON, NIGHT");
                    return;
                }
                createShift(args[2], args[3], args[4]);
                break;
            case "publish":
                if (args.length < 3) {
                    System.out.println("用法: shift publish <shiftId>");
                    return;
                }
                publishShift(args[2]);
                break;
            case "publish-batch":
                if (args.length < 3) {
                    System.out.println("用法: shift publish-batch <shiftId1,shiftId2...>");
                    return;
                }
                publishShiftsBatch(args[2]);
                break;
            case "delete":
                if (args.length < 3) {
                    System.out.println("用法: shift delete <shiftId>");
                    return;
                }
                deleteShift(args[2]);
                break;
            default:
                System.out.println("未知操作: " + action);
                printHelp();
        }
    }

    private void listAllShifts() throws Exception {
        TypeReference<ApiResponse<List<ShiftDTO>>> typeRef = 
                new TypeReference<ApiResponse<List<ShiftDTO>>>() {};
        
        LocalDate today = LocalDate.now();
        LocalDate start = today.minusDays(30);
        LocalDate end = today.plusDays(30);
        
        Map<String, String> params = new HashMap<>();
        params.put("startDate", start.toString());
        params.put("endDate", end.toString());
        
        String path = httpClient.buildPathWithParams("/api/shifts/range", params);
        ApiResponse<List<ShiftDTO>> response = httpClient.get(path, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printShifts(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void listShiftsByEmployee(String employeeId) throws Exception {
        TypeReference<ApiResponse<List<ShiftDTO>>> typeRef = 
                new TypeReference<ApiResponse<List<ShiftDTO>>>() {};
        ApiResponse<List<ShiftDTO>> response = httpClient.get("/api/shifts/employee/" + employeeId, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printShifts(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void listShiftsByDateRange(String startDateStr, String endDateStr) throws Exception {
        LocalDate startDate = LocalDate.parse(startDateStr);
        LocalDate endDate = LocalDate.parse(endDateStr);
        
        if (startDate.isAfter(endDate)) {
            OutputFormatter.printError("开始日期不能晚于结束日期");
            return;
        }

        TypeReference<ApiResponse<List<ShiftDTO>>> typeRef = 
                new TypeReference<ApiResponse<List<ShiftDTO>>>() {};
        
        Map<String, String> params = new HashMap<>();
        params.put("startDate", startDate.toString());
        params.put("endDate", endDate.toString());
        
        String path = httpClient.buildPathWithParams("/api/shifts/range", params);
        ApiResponse<List<ShiftDTO>> response = httpClient.get(path, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printShifts(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void createShift(String employeeId, String dateStr, String shiftTypeStr) throws Exception {
        ShiftType shiftType;
        try {
            shiftType = ShiftType.valueOf(shiftTypeStr.toUpperCase());
        } catch (IllegalArgumentException e) {
            OutputFormatter.printError("无效的班次类型: " + shiftTypeStr);
            System.out.println("可用类型: MORNING, AFTERNOON, NIGHT");
            return;
        }

        LocalDate date;
        try {
            date = LocalDate.parse(dateStr);
        } catch (Exception e) {
            OutputFormatter.printError("无效的日期格式: " + dateStr);
            System.out.println("日期格式应为: yyyy-MM-dd");
            return;
        }

        CreateShiftRequest request = new CreateShiftRequest();
        request.setEmployeeId(employeeId);
        request.setDate(date);
        request.setShiftType(shiftType);

        TypeReference<ApiResponse<ShiftDTO>> typeRef = 
                new TypeReference<ApiResponse<ShiftDTO>>() {};
        ApiResponse<ShiftDTO> response = httpClient.post("/api/shifts", request, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSuccess("排班创建成功");
            OutputFormatter.printShift(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void publishShift(String shiftId) throws Exception {
        TypeReference<ApiResponse<Void>> typeRef = 
                new TypeReference<ApiResponse<Void>>() {};
        ApiResponse<Void> response = httpClient.post("/api/shifts/" + shiftId + "/publish", null, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSuccess("排班发布成功");
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void publishShiftsBatch(String idsStr) throws Exception {
        List<String> shiftIds = Arrays.asList(idsStr.split(","));
        
        TypeReference<ApiResponse<Void>> typeRef = 
                new TypeReference<ApiResponse<Void>>() {};
        ApiResponse<Void> response = httpClient.post("/api/shifts/publish-batch", shiftIds, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSuccess("批量发布成功，共 " + shiftIds.size() + " 条排班");
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void deleteShift(String shiftId) throws Exception {
        TypeReference<ApiResponse<Void>> typeRef = 
                new TypeReference<ApiResponse<Void>>() {};
        ApiResponse<Void> response = httpClient.delete("/api/shifts/" + shiftId, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSuccess("排班删除成功");
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void printHelp() {
        System.out.println("排班管理命令:");
        System.out.println("  list                                  - 列出最近60天的排班");
        System.out.println("  list <employeeId>                     - 列出指定员工的排班");
        System.out.println("  range <startDate> <endDate>          - 列出指定日期范围的排班");
        System.out.println("  create <employeeId> <date> <type>    - 创建排班");
        System.out.println("  publish <shiftId>                     - 发布排班");
        System.out.println("  publish-batch <shiftId1,shiftId2...> - 批量发布排班");
        System.out.println("  delete <shiftId>                      - 删除排班");
        System.out.println();
        System.out.println("班次类型: MORNING(早班), AFTERNOON(中班), NIGHT(夜班)");
        System.out.println("日期格式: yyyy-MM-dd");
    }
}