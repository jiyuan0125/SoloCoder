package com.employee.common.dto;

public class DepartmentStatsDTO {
    private String department;
    private int totalCount;
    private int activeCount;
    private int resignedCount;

    public String getDepartment() {
        return department;
    }

    public void setDepartment(String department) {
        this.department = department;
    }

    public int getTotalCount() {
        return totalCount;
    }

    public void setTotalCount(int totalCount) {
        this.totalCount = totalCount;
    }

    public int getActiveCount() {
        return activeCount;
    }

    public void setActiveCount(int activeCount) {
        this.activeCount = activeCount;
    }

    public int getResignedCount() {
        return resignedCount;
    }

    public void setResignedCount(int resignedCount) {
        this.resignedCount = resignedCount;
    }
}
