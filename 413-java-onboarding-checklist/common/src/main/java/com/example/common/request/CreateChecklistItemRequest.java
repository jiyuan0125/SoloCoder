package com.example.common.request;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import java.time.LocalDate;

public class CreateChecklistItemRequest {
    @NotBlank(message = "事项名称不能为空")
    private String name;
    private String description;
    @NotBlank(message = "负责人不能为空")
    private String responsiblePerson;
    private String department;
    @NotNull(message = "截止日期不能为空")
    private LocalDate dueDate;
    private boolean isRequired;
    private boolean isOnboardingDayRequired;
    private boolean isPreOnboarding;

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
}
