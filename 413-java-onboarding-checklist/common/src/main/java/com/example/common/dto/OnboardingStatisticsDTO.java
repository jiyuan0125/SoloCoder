package com.example.common.dto;

import java.util.Map;

public class OnboardingStatisticsDTO {
    private long totalEmployees;
    private long completedOnboarding;
    private long incompleteOnboarding;
    private double completionRate;
    private Map<String, Long> itemsByStatus;
    private Map<String, Long> overdueItemsByResponsible;
    private Map<String, Long> employeesByPosition;
    private double averageCompletionTime;

    public long getTotalEmployees() {
        return totalEmployees;
    }

    public void setTotalEmployees(long totalEmployees) {
        this.totalEmployees = totalEmployees;
    }

    public long getCompletedOnboarding() {
        return completedOnboarding;
    }

    public void setCompletedOnboarding(long completedOnboarding) {
        this.completedOnboarding = completedOnboarding;
    }

    public long getIncompleteOnboarding() {
        return incompleteOnboarding;
    }

    public void setIncompleteOnboarding(long incompleteOnboarding) {
        this.incompleteOnboarding = incompleteOnboarding;
    }

    public double getCompletionRate() {
        return completionRate;
    }

    public void setCompletionRate(double completionRate) {
        this.completionRate = completionRate;
    }

    public Map<String, Long> getItemsByStatus() {
        return itemsByStatus;
    }

    public void setItemsByStatus(Map<String, Long> itemsByStatus) {
        this.itemsByStatus = itemsByStatus;
    }

    public Map<String, Long> getOverdueItemsByResponsible() {
        return overdueItemsByResponsible;
    }

    public void setOverdueItemsByResponsible(Map<String, Long> overdueItemsByResponsible) {
        this.overdueItemsByResponsible = overdueItemsByResponsible;
    }

    public Map<String, Long> getEmployeesByPosition() {
        return employeesByPosition;
    }

    public void setEmployeesByPosition(Map<String, Long> employeesByPosition) {
        this.employeesByPosition = employeesByPosition;
    }

    public double getAverageCompletionTime() {
        return averageCompletionTime;
    }

    public void setAverageCompletionTime(double averageCompletionTime) {
        this.averageCompletionTime = averageCompletionTime;
    }
}
