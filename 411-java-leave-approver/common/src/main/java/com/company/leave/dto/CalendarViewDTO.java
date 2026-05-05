package com.company.leave.dto;

import com.company.leave.enums.LeaveStatus;
import com.company.leave.enums.LeaveType;

import java.time.LocalDate;
import java.util.ArrayList;
import java.util.List;

public class CalendarViewDTO {
    private int year;
    private int month;
    private List<DayLeaveStatus> dayStatuses;

    public CalendarViewDTO() {
        this.dayStatuses = new ArrayList<>();
    }

    public int getYear() {
        return year;
    }

    public void setYear(int year) {
        this.year = year;
    }

    public int getMonth() {
        return month;
    }

    public void setMonth(int month) {
        this.month = month;
    }

    public List<DayLeaveStatus> getDayStatuses() {
        return dayStatuses;
    }

    public void setDayStatuses(List<DayLeaveStatus> dayStatuses) {
        this.dayStatuses = dayStatuses;
    }

    public static class DayLeaveStatus {
        private LocalDate date;
        private List<EmployeeLeave> leaves;

        public DayLeaveStatus() {
            this.leaves = new ArrayList<>();
        }

        public LocalDate getDate() {
            return date;
        }

        public void setDate(LocalDate date) {
            this.date = date;
        }

        public List<EmployeeLeave> getLeaves() {
            return leaves;
        }

        public void setLeaves(List<EmployeeLeave> leaves) {
            this.leaves = leaves;
        }
    }

    public static class EmployeeLeave {
        private Long employeeId;
        private String employeeName;
        private LeaveType leaveType;
        private LeaveStatus status;

        public EmployeeLeave() {
        }

        public Long getEmployeeId() {
            return employeeId;
        }

        public void setEmployeeId(Long employeeId) {
            this.employeeId = employeeId;
        }

        public String getEmployeeName() {
            return employeeName;
        }

        public void setEmployeeName(String employeeName) {
            this.employeeName = employeeName;
        }

        public LeaveType getLeaveType() {
            return leaveType;
        }

        public void setLeaveType(LeaveType leaveType) {
            this.leaveType = leaveType;
        }

        public LeaveStatus getStatus() {
            return status;
        }

        public void setStatus(LeaveStatus status) {
            this.status = status;
        }
    }
}
