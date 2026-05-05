package com.company.leave.service;

import com.company.leave.dto.CalendarViewDTO;
import com.company.leave.entity.Employee;
import com.company.leave.entity.LeaveRecord;
import com.company.leave.enums.LeaveType;
import org.springframework.stereotype.Service;

import java.time.LocalDate;
import java.time.YearMonth;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Service
public class CalendarService {
    private final LeaveService leaveService;
    private final EmployeeService employeeService;

    public CalendarService(LeaveService leaveService, EmployeeService employeeService) {
        this.leaveService = leaveService;
        this.employeeService = employeeService;
    }

    public CalendarViewDTO getTeamCalendar(Long managerId, int year, int month) {
        List<Employee> teamMembers = employeeService.getTeamMembers(managerId);
        
        YearMonth yearMonth = YearMonth.of(year, month);
        LocalDate startDate = yearMonth.atDay(1);
        LocalDate endDate = yearMonth.atEndOfMonth();
        
        List<LeaveRecord> allLeaves = leaveService.getLeaveRecordsByDateRange(startDate, endDate);
        
        Map<Long, Employee> employeeMap = teamMembers.stream()
                .collect(Collectors.toMap(Employee::getId, e -> e));
        
        List<LeaveRecord> teamLeaves = allLeaves.stream()
                .filter(l -> employeeMap.containsKey(l.getEmployeeId()))
                .collect(Collectors.toList());
        
        CalendarViewDTO calendarView = new CalendarViewDTO();
        calendarView.setYear(year);
        calendarView.setMonth(month);
        
        List<CalendarViewDTO.DayLeaveStatus> dayStatuses = new ArrayList<>();
        LocalDate current = startDate;
        
        while (!current.isAfter(endDate)) {
            CalendarViewDTO.DayLeaveStatus dayStatus = new CalendarViewDTO.DayLeaveStatus();
            dayStatus.setDate(current);
            
            List<CalendarViewDTO.EmployeeLeave> leavesOnDay = new ArrayList<>();
            for (LeaveRecord leave : teamLeaves) {
                if (isDateInRange(current, leave.getStartDate(), leave.getEndDate())) {
                    if (isCountedAsLeaveDayForCalendar(current, leave.getLeaveType())) {
                        Employee employee = employeeMap.get(leave.getEmployeeId());
                        CalendarViewDTO.EmployeeLeave empLeave = new CalendarViewDTO.EmployeeLeave();
                        empLeave.setEmployeeId(employee.getId());
                        empLeave.setEmployeeName(employee.getName());
                        empLeave.setLeaveType(leave.getLeaveType());
                        empLeave.setStatus(leave.getStatus());
                        leavesOnDay.add(empLeave);
                    }
                }
            }
            
            dayStatus.setLeaves(leavesOnDay);
            dayStatuses.add(dayStatus);
            current = current.plusDays(1);
        }
        
        calendarView.setDayStatuses(dayStatuses);
        return calendarView;
    }

    private boolean isDateInRange(LocalDate date, LocalDate start, LocalDate end) {
        return !date.isBefore(start) && !date.isAfter(end);
    }

    private boolean isCountedAsLeaveDayForCalendar(LocalDate date, LeaveType leaveType) {
        if (leaveType == null) {
            return true;
        }
        if (isHoliday(date)) {
            return false;
        }
        if (isWeekend(date)) {
            return leaveType.isCountWeekends();
        }
        return true;
    }

    private boolean isHoliday(LocalDate date) {
        return com.company.leave.util.DateCalculator.isHoliday(date);
    }

    private boolean isWeekend(LocalDate date) {
        return com.company.leave.util.DateCalculator.isWeekend(date);
    }
}
