package com.workshift.common.request;

import com.workshift.common.enums.ShiftType;
import java.time.LocalDate;

public class CreateShiftRequest {
    private String employeeId;
    private LocalDate date;
    private ShiftType shiftType;

    public CreateShiftRequest() {
    }

    public String getEmployeeId() {
        return employeeId;
    }

    public void setEmployeeId(String employeeId) {
        this.employeeId = employeeId;
    }

    public LocalDate getDate() {
        return date;
    }

    public void setDate(LocalDate date) {
        this.date = date;
    }

    public ShiftType getShiftType() {
        return shiftType;
    }

    public void setShiftType(ShiftType shiftType) {
        this.shiftType = shiftType;
    }
}