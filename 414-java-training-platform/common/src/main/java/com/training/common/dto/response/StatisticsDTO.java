package com.training.common.dto.response;

import java.util.List;
import java.util.Map;

public class StatisticsDTO {
    private int totalCourses;
    private int publishedCourses;
    private int ongoingCourses;
    private int completedCourses;
    private int totalEmployees;
    private int totalRegistrations;
    private int completedRegistrations;
    private int passedRegistrations;
    private int failedRegistrations;
    private double passRate;
    private int requiredCourses;
    private int electiveCourses;
    private List<DepartmentStatisticsDTO> departmentStatistics;
    private Map<String, Integer> courseTypeDistribution;
    private List<MonthlyStatisticsDTO> monthlyStatistics;

    public StatisticsDTO() {
    }

    public int getTotalCourses() {
        return totalCourses;
    }

    public void setTotalCourses(int totalCourses) {
        this.totalCourses = totalCourses;
    }

    public int getPublishedCourses() {
        return publishedCourses;
    }

    public void setPublishedCourses(int publishedCourses) {
        this.publishedCourses = publishedCourses;
    }

    public int getOngoingCourses() {
        return ongoingCourses;
    }

    public void setOngoingCourses(int ongoingCourses) {
        this.ongoingCourses = ongoingCourses;
    }

    public int getCompletedCourses() {
        return completedCourses;
    }

    public void setCompletedCourses(int completedCourses) {
        this.completedCourses = completedCourses;
    }

    public int getTotalEmployees() {
        return totalEmployees;
    }

    public void setTotalEmployees(int totalEmployees) {
        this.totalEmployees = totalEmployees;
    }

    public int getTotalRegistrations() {
        return totalRegistrations;
    }

    public void setTotalRegistrations(int totalRegistrations) {
        this.totalRegistrations = totalRegistrations;
    }

    public int getCompletedRegistrations() {
        return completedRegistrations;
    }

    public void setCompletedRegistrations(int completedRegistrations) {
        this.completedRegistrations = completedRegistrations;
    }

    public int getPassedRegistrations() {
        return passedRegistrations;
    }

    public void setPassedRegistrations(int passedRegistrations) {
        this.passedRegistrations = passedRegistrations;
    }

    public int getFailedRegistrations() {
        return failedRegistrations;
    }

    public void setFailedRegistrations(int failedRegistrations) {
        this.failedRegistrations = failedRegistrations;
    }

    public double getPassRate() {
        return passRate;
    }

    public void setPassRate(double passRate) {
        this.passRate = passRate;
    }

    public int getRequiredCourses() {
        return requiredCourses;
    }

    public void setRequiredCourses(int requiredCourses) {
        this.requiredCourses = requiredCourses;
    }

    public int getElectiveCourses() {
        return electiveCourses;
    }

    public void setElectiveCourses(int electiveCourses) {
        this.electiveCourses = electiveCourses;
    }

    public List<DepartmentStatisticsDTO> getDepartmentStatistics() {
        return departmentStatistics;
    }

    public void setDepartmentStatistics(List<DepartmentStatisticsDTO> departmentStatistics) {
        this.departmentStatistics = departmentStatistics;
    }

    public Map<String, Integer> getCourseTypeDistribution() {
        return courseTypeDistribution;
    }

    public void setCourseTypeDistribution(Map<String, Integer> courseTypeDistribution) {
        this.courseTypeDistribution = courseTypeDistribution;
    }

    public List<MonthlyStatisticsDTO> getMonthlyStatistics() {
        return monthlyStatistics;
    }

    public void setMonthlyStatistics(List<MonthlyStatisticsDTO> monthlyStatistics) {
        this.monthlyStatistics = monthlyStatistics;
    }
}
