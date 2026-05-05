package com.example.common.dto;

import com.example.common.enums.OnboardingStatus;
import com.example.common.enums.PositionType;
import java.time.LocalDate;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

public class EmployeeDTO {
    private String id;
    private String name;
    private String email;
    private PositionType positionType;
    private String department;
    private LocalDate onboardingDate;
    private OnboardingStatus status;
    private List<ChecklistItemDTO> checklistItems;

    public EmployeeDTO() {
        this.id = UUID.randomUUID().toString();
        this.checklistItems = new ArrayList<>();
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

    public String getEmail() {
        return email;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public PositionType getPositionType() {
        return positionType;
    }

    public void setPositionType(PositionType positionType) {
        this.positionType = positionType;
    }

    public String getDepartment() {
        return department;
    }

    public void setDepartment(String department) {
        this.department = department;
    }

    public LocalDate getOnboardingDate() {
        return onboardingDate;
    }

    public void setOnboardingDate(LocalDate onboardingDate) {
        this.onboardingDate = onboardingDate;
    }

    public OnboardingStatus getStatus() {
        return status;
    }

    public void setStatus(OnboardingStatus status) {
        this.status = status;
    }

    public List<ChecklistItemDTO> getChecklistItems() {
        return checklistItems;
    }

    public void setChecklistItems(List<ChecklistItemDTO> checklistItems) {
        this.checklistItems = checklistItems;
    }
}
