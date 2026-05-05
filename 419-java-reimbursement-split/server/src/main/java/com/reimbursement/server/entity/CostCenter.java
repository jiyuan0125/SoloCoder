package com.reimbursement.server.entity;

import java.math.BigDecimal;

public class CostCenter {
    private String id;
    private String name;
    private String managerId;
    private String managerName;
    private BigDecimal monthlyBudget;

    public CostCenter() {
    }

    public CostCenter(String id, String name, String managerId, String managerName, BigDecimal monthlyBudget) {
        this.id = id;
        this.name = name;
        this.managerId = managerId;
        this.managerName = managerName;
        this.monthlyBudget = monthlyBudget;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getManagerId() {
        return managerId;
    }

    public void setManagerId(String managerId) {
        this.managerId = managerId;
    }

    public String getManagerName() {
        return managerName;
    }

    public void setManagerName(String managerName) {
        this.managerName = managerName;
    }

    public BigDecimal getMonthlyBudget() {
        return monthlyBudget;
    }

    public void setMonthlyBudget(BigDecimal monthlyBudget) {
        this.monthlyBudget = monthlyBudget;
    }
}
