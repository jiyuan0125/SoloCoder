package com.recruitment.common.dto;

import com.recruitment.common.enums.OfferStatus;

import java.math.BigDecimal;
import java.time.LocalDateTime;

public class OfferDTO {
    private String id;
    private String applicationId;
    private BigDecimal salary;
    private LocalDateTime sentTime;
    private LocalDateTime expireTime;
    private LocalDateTime responseTime;
    private OfferStatus status;
    private boolean needReapproval;

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

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

    public LocalDateTime getSentTime() {
        return sentTime;
    }

    public void setSentTime(LocalDateTime sentTime) {
        this.sentTime = sentTime;
    }

    public LocalDateTime getExpireTime() {
        return expireTime;
    }

    public void setExpireTime(LocalDateTime expireTime) {
        this.expireTime = expireTime;
    }

    public LocalDateTime getResponseTime() {
        return responseTime;
    }

    public void setResponseTime(LocalDateTime responseTime) {
        this.responseTime = responseTime;
    }

    public OfferStatus getStatus() {
        return status;
    }

    public void setStatus(OfferStatus status) {
        this.status = status;
    }

    public boolean isNeedReapproval() {
        return needReapproval;
    }

    public void setNeedReapproval(boolean needReapproval) {
        this.needReapproval = needReapproval;
    }
}
