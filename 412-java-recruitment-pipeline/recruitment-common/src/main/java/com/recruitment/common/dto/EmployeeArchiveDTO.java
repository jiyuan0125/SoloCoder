package com.recruitment.common.dto;

import com.recruitment.common.enums.SourceChannel;

import java.time.LocalDateTime;

public class EmployeeArchiveDTO {
    private String id;
    private String candidateId;
    private String name;
    private String phone;
    private String email;
    private String position;
    private java.math.BigDecimal salary;
    private SourceChannel sourceChannel;
    private LocalDateTime onboardDate;
    private LocalDateTime createdAt;

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getCandidateId() {
        return candidateId;
    }

    public void setCandidateId(String candidateId) {
        this.candidateId = candidateId;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getPhone() {
        return phone;
    }

    public void setPhone(String phone) {
        this.phone = phone;
    }

    public String getEmail() {
        return email;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public String getPosition() {
        return position;
    }

    public void setPosition(String position) {
        this.position = position;
    }

    public java.math.BigDecimal getSalary() {
        return salary;
    }

    public void setSalary(java.math.BigDecimal salary) {
        this.salary = salary;
    }

    public SourceChannel getSourceChannel() {
        return sourceChannel;
    }

    public void setSourceChannel(SourceChannel sourceChannel) {
        this.sourceChannel = sourceChannel;
    }

    public LocalDateTime getOnboardDate() {
        return onboardDate;
    }

    public void setOnboardDate(LocalDateTime onboardDate) {
        this.onboardDate = onboardDate;
    }

    public LocalDateTime getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(LocalDateTime createdAt) {
        this.createdAt = createdAt;
    }
}
