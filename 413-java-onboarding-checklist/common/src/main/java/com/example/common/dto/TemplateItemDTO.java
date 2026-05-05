package com.example.common.dto;

import java.util.UUID;

public class TemplateItemDTO {
    private String id;
    private String name;
    private String description;
    private String responsiblePerson;
    private String department;
    private int daysRelativeToOnboarding;
    private boolean isRequired;
    private boolean isOnboardingDayRequired;
    private boolean isPreOnboarding;

    public TemplateItemDTO() {
        this.id = UUID.randomUUID().toString();
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

    public int getDaysRelativeToOnboarding() {
        return daysRelativeToOnboarding;
    }

    public void setDaysRelativeToOnboarding(int daysRelativeToOnboarding) {
        this.daysRelativeToOnboarding = daysRelativeToOnboarding;
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
