package com.workshift.common.dto;

import java.time.LocalDate;

public class HolidayDTO {
    private LocalDate date;
    private String name;
    private boolean isStatutory;

    public HolidayDTO() {
    }

    public HolidayDTO(LocalDate date, String name, boolean isStatutory) {
        this.date = date;
        this.name = name;
        this.isStatutory = isStatutory;
    }

    public LocalDate getDate() {
        return date;
    }

    public void setDate(LocalDate date) {
        this.date = date;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public boolean isStatutory() {
        return isStatutory;
    }

    public void setStatutory(boolean statutory) {
        isStatutory = statutory;
    }
}