package com.workshift.server.controller;

import com.workshift.common.dto.HolidayDTO;
import com.workshift.common.dto.MonthlyStatisticsDTO;
import com.workshift.common.enums.ErrorCode;
import com.workshift.common.response.ApiResponse;
import com.workshift.server.service.StatisticsService;
import java.time.LocalDate;
import java.time.YearMonth;
import java.util.List;
import java.util.Optional;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/statistics")
public class StatisticsController {

    private final StatisticsService statisticsService;

    public StatisticsController(StatisticsService statisticsService) {
        this.statisticsService = statisticsService;
    }

    @GetMapping("/employee/{employeeId}/monthly")
    public ApiResponse<MonthlyStatisticsDTO> getMonthlyStatistics(
            @PathVariable String employeeId,
            @RequestParam int year,
            @RequestParam int month) {
        
        if (year < 2000 || year > 2100 || month < 1 || month > 12) {
            return ApiResponse.error(ErrorCode.INVALID_PARAMETER.getCode(), "无效的年月参数");
        }

        YearMonth yearMonth = YearMonth.of(year, month);
        MonthlyStatisticsDTO stats = statisticsService.generateMonthlyStatistics(employeeId, yearMonth);
        
        if (stats == null) {
            return ApiResponse.error(ErrorCode.EMPLOYEE_NOT_FOUND.getCode(), ErrorCode.EMPLOYEE_NOT_FOUND.getMessage());
        }
        return ApiResponse.success(stats);
    }

    @GetMapping("/all/monthly")
    public ApiResponse<List<MonthlyStatisticsDTO>> getAllMonthlyStatistics(
            @RequestParam int year,
            @RequestParam int month) {
        
        if (year < 2000 || year > 2100 || month < 1 || month > 12) {
            return ApiResponse.error(ErrorCode.INVALID_PARAMETER.getCode(), "无效的年月参数");
        }

        YearMonth yearMonth = YearMonth.of(year, month);
        List<MonthlyStatisticsDTO> stats = statisticsService.generateAllMonthlyStatistics(yearMonth);
        return ApiResponse.success(stats);
    }

    @PostMapping("/holidays")
    public ApiResponse<Void> addHoliday(@RequestBody HolidayDTO holiday) {
        if (holiday.getDate() == null || holiday.getName() == null) {
            return ApiResponse.error(ErrorCode.INVALID_PARAMETER.getCode(), "参数不完整");
        }

        Optional<ErrorCode> result = statisticsService.addHoliday(holiday.getDate(), holiday.getName());
        if (result.isPresent()) {
            return ApiResponse.error(result.get().getCode(), result.get().getMessage());
        }
        return ApiResponse.success();
    }

    @GetMapping("/holidays/check")
    public ApiResponse<Boolean> checkHoliday(
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate date) {
        boolean isHoliday = statisticsService.isHoliday(date);
        return ApiResponse.success(isHoliday);
    }
}