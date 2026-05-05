package com.company.leave.entity;

import java.time.LocalDate;

public class Employee {
    private Long id;
    private String name;
    private Long managerId;
    private LocalDate joinDate;
    private int annualLeaveQuota;
    private int annualLeaveRemaining;
    private int carriedOverLeave;

    public Employee() {
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public Long getManagerId() {
        return managerId;
    }

    public void setManagerId(Long managerId) {
        this.managerId = managerId;
    }

    public LocalDate getJoinDate() {
        return joinDate;
    }

    public void setJoinDate(LocalDate joinDate) {
        this.joinDate = joinDate;
    }

    public int getAnnualLeaveQuota() {
        return annualLeaveQuota;
    }

    public void setAnnualLeaveQuota(int annualLeaveQuota) {
        this.annualLeaveQuota = annualLeaveQuota;
    }

    public int getAnnualLeaveRemaining() {
        return annualLeaveRemaining;
    }

    public void setAnnualLeaveRemaining(int annualLeaveRemaining) {
        this.annualLeaveRemaining = annualLeaveRemaining;
    }

    public int getCarriedOverLeave() {
        return carriedOverLeave;
    }

    public void setCarriedOverLeave(int carriedOverLeave) {
        this.carriedOverLeave = carriedOverLeave;
    }
}
