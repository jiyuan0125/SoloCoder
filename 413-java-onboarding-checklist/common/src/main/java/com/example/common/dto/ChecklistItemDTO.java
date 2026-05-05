package com.example.common.dto;

import com.example.common.enums.ItemStatus;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.UUID;

public class ChecklistItemDTO {
    private String id;
    private String name;
    private String description;
    private String responsiblePerson;
    private String department;
    private LocalDate dueDate;
    private LocalDate completedDate;
    private ItemStatus status;
    private boolean isRequired;
    private boolean isOnboardingDayRequired;
    private boolean isPreOnboarding;
    private boolean isEscalated;
    private String escalatedTo;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

    public ChecklistItemDTO() {
        this.id = UUID.randomUUID().toString();
        this.status = ItemStatus.PENDING;
        this.createdAt = LocalDateTime.now();
        this.updatedAt = LocalDateTime.now();
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

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public String getResponsiblePerson() {
        return responsiblePerson;
    }

    public void setResponsiblePerson(String responsiblePerson) {
        this.responsiblePerson = responsiblePerson;
    }

    public String getDepartment() {
        return department;
    }

    public void setDepartment(String department) {
        this.department = department;
    }

    public LocalDate getDueDate() {
        return dueDate;
    }

    public void setDueDate(LocalDate dueDate) {
        this.dueDate = dueDate;
    }

    public LocalDate getCompletedDate() {
        return completedDate;
    }

    public void setCompletedDate(LocalDate completedDate) {
        this.completedDate = completedDate;
    }

    public ItemStatus getStatus() {
        return status;
    }

    public void setStatus(ItemStatus status) {
        this.status = status;
    }

    public boolean isRequired() {
        return isRequired;
    }

    public void setRequired(boolean required) {
        isRequired = required;
    }

    public boolean isOnboardingDayRequired() {
        return isOnboardingDayRequired;
    }

    public void setOnboardingDayRequired(boolean onboardingDayRequired) {
        isOnboardingDayRequired = onboardingDayRequired;
    }

    public boolean isPreOnboarding() {
        return isPreOnboarding;
    }

    public void setPreOnboarding(boolean preOnboarding) {
        isPreOnboarding = preOnboarding;
    }

    public boolean isEscalated() {
        return isEscalated;
    }

    public void setEscalated(boolean escalated) {
        isEscalated = escalated;
    }

    public String getEscalatedTo() {
        return escalatedTo;
    }

    public void setEscalatedTo(String escalatedTo) {
        this.escalatedTo = escalatedTo;
    }

    public LocalDateTime getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(LocalDateTime createdAt) {
        this.createdAt = createdAt;
    }

    public LocalDateTime getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(LocalDateTime updatedAt) {
        this.updatedAt = updatedAt;
    }
}
