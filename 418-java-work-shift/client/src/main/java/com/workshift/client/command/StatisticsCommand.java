package com.workshift.client.command;

import com.fasterxml.jackson.core.type.TypeReference;
import com.workshift.common.dto.HolidayDTO;
import com.workshift.common.dto.MonthlyStatisticsDTO;
import com.workshift.common.response.ApiResponse;
import com.workshift.client.http.HttpClient;
import com.workshift.client.util.OutputFormatter;
import java.time.LocalDate;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class StatisticsCommand implements Command {

    private final HttpClient httpClient;

    public StatisticsCommand(HttpClient httpClient) {
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
            case "employee":
                if (args.length < 5) {
                    System.out.println("用法: stats employee <employeeId> <year> <month>");
                    return;
                }
                getEmployeeStatistics(args[2], args[3], args[4]);
                break;
            case "all":
                if (args.length < 4) {
                    System.out.println("用法: stats all <year> <month>");
                    return;
                }
                getAllStatistics(args[2], args[3]);
                break;
            case "holiday":
                if (args.length < 3) {
                    System.out.println("用法: stats holiday add <date> <name>");
                    System.out.println("       stats holiday check <date>");
                    return;
                }
                handleHoliday(args);
                break;
            default:
                System.out.println("未知操作: " + action);
                printHelp();
        }
    }

    private void getEmployeeStatistics(String employeeId, String yearStr, String monthStr) throws Exception {
        int year;
        int month;
        try {
            year = Integer.parseInt(yearStr);
            month = Integer.parseInt(monthStr);
        } catch (NumberFormatException e) {
            OutputFormatter.printError("无效的年份或月份");
            return;
        }

        Map<String, String> params = new HashMap<>();
        params.put("year", String.valueOf(year));
        params.put("month", String.valueOf(month));

        TypeReference<ApiResponse<MonthlyStatisticsDTO>> typeRef = 
                new TypeReference<ApiResponse<MonthlyStatisticsDTO>>() {};
        String path = httpClient.buildPathWithParams("/api/statistics/employee/" + employeeId + "/monthly", params);
        ApiResponse<MonthlyStatisticsDTO> response = httpClient.get(path, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printMonthlyStatistics(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void getAllStatistics(String yearStr, String monthStr) throws Exception {
        int year;
        int month;
        try {
            year = Integer.parseInt(yearStr);
            month = Integer.parseInt(monthStr);
        } catch (NumberFormatException e) {
            OutputFormatter.printError("无效的年份或月份");
            return;
        }

        Map<String, String> params = new HashMap<>();
        params.put("year", String.valueOf(year));
        params.put("month", String.valueOf(month));

        TypeReference<ApiResponse<List<MonthlyStatisticsDTO>>> typeRef = 
                new TypeReference<ApiResponse<List<MonthlyStatisticsDTO>>>() {};
        String path = httpClient.buildPathWithParams("/api/statistics/all/monthly", params);
        ApiResponse<List<MonthlyStatisticsDTO>> response = httpClient.get(path, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printAllMonthlyStatistics(response.getData());
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void handleHoliday(String[] args) throws Exception {
        String subAction = args[2];

        switch (subAction.toLowerCase()) {
            case "add":
                if (args.length < 5) {
                    System.out.println("用法: stats holiday add <date> <name>");
                    return;
                }
                addHoliday(args[3], args, 4);
                break;
            case "check":
                if (args.length < 4) {
                    System.out.println("用法: stats holiday check <date>");
                    return;
                }
                checkHoliday(args[3]);
                break;
            default:
                System.out.println("未知操作: " + subAction);
                System.out.println("可用操作: add, check");
        }
    }

    private void addHoliday(String dateStr, String[] args, int startIndex) throws Exception {
        LocalDate date;
        try {
            date = LocalDate.parse(dateStr);
        } catch (Exception e) {
            OutputFormatter.printError("无效的日期格式: " + dateStr);
            System.out.println("日期格式应为: yyyy-MM-dd");
            return;
        }

        StringBuilder name = new StringBuilder();
        for (int i = startIndex; i < args.length; i++) {
            if (i > startIndex) name.append(" ");
            name.append(args[i]);
        }

        HolidayDTO holiday = new HolidayDTO();
        holiday.setDate(date);
        holiday.setName(name.toString());
        holiday.setStatutory(true);

        TypeReference<ApiResponse<Void>> typeRef = 
                new TypeReference<ApiResponse<Void>>() {};
        ApiResponse<Void> response = httpClient.post("/api/statistics/holidays", holiday, typeRef);
        
        if (response.getCode() == 0) {
            OutputFormatter.printSuccess("节假日添加成功: " + date + " - " + name);
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void checkHoliday(String dateStr) throws Exception {
        LocalDate date;
        try {
            date = LocalDate.parse(dateStr);
        } catch (Exception e) {
            OutputFormatter.printError("无效的日期格式: " + dateStr);
            System.out.println("日期格式应为: yyyy-MM-dd");
            return;
        }

        Map<String, String> params = new HashMap<>();
        params.put("date", date.toString());

        TypeReference<ApiResponse<Boolean>> typeRef = 
                new TypeReference<ApiResponse<Boolean>>() {};
        String path = httpClient.buildPathWithParams("/api/statistics/holidays/check", params);
        ApiResponse<Boolean> response = httpClient.get(path, typeRef);
        
        if (response.getCode() == 0) {
            if (response.getData()) {
                System.out.println("日期 " + date + " 是节假日");
            } else {
                System.out.println("日期 " + date + " 不是节假日");
            }
        } else {
            OutputFormatter.printError(response.getMessage());
        }
    }

    private void printHelp() {
        System.out.println("统计报表命令:");
        System.out.println("  employee <employeeId> <year> <month>  - 生成员工月度统计");
        System.out.println("  all <year> <month>                     - 生成所有员工月度统计");
        System.out.println("  holiday add <date> <name>              - 添加法定节假日");
        System.out.println("  holiday check <date>                   - 检查日期是否为节假日");
        System.out.println();
        System.out.println("日期格式: yyyy-MM-dd");
    }
}