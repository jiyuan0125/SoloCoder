package com.recruitment.common.request;

import java.math.BigDecimal;

public class SendOfferRequest {
    private String applicationId;
    private BigDecimal salary;

    public String getApplicationId() {
        return applicationId;
    }

    public void setApplicationId(String applicationId) {
        this.applicationId = applicationId;
    }

    public BigDecimal getSalary() {
        return salary;
    }

    public void setSalary(BigDecimal salary) {
        this.salary = salary;
    }
}
