package com.company.leave.controller;

import com.company.leave.dto.ApiResponse;
import com.company.leave.dto.CalendarViewDTO;
import com.company.leave.service.CalendarService;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/calendar")
public class CalendarController {
    private final CalendarService calendarService;

    public CalendarController(CalendarService calendarService) {
        this.calendarService = calendarService;
    }

    @GetMapping("/team/{managerId}")
    public ApiResponse<CalendarViewDTO> getTeamCalendar(
            @PathVariable Long managerId,
            @RequestParam int year,
            @RequestParam int month) {
        CalendarViewDTO calendar = calendarService.getTeamCalendar(managerId, year, month);
        return ApiResponse.success(calendar);
    }
}
