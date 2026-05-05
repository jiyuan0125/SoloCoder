package com.training.common.dto.response;

public class MonthlyStatisticsDTO {
    private int year;
    private int month;
    private int coursesStarted;
    private int coursesCompleted;
    private int registrations;
    private int completions;
    private int creditsEarned;

    public MonthlyStatisticsDTO() {
    }

    public int getYear() {
        return year;
    }

    public void setYear(int year) {
        this.year = year;
    }

    public int getMonth() {
        return month;
    }

    public void setMonth(int month) {
        this.month = month;
    }

    public int getCoursesStarted() {
        return coursesStarted;
    }

    public void setCoursesStarted(int coursesStarted) {
        this.coursesStarted = coursesStarted;
    }

    public int getCoursesCompleted() {
        return coursesCompleted;
    }

    public void setCoursesCompleted(int coursesCompleted) {
        this.coursesCompleted = coursesCompleted;
    }

    public int getRegistrations() {
        return registrations;
    }

    public void setRegistrations(int registrations) {
        this.registrations = registrations;
    }

    public int getCompletions() {
        return completions;
    }

    public void setCompletions(int completions) {
        this.completions = completions;
    }

    public int getCreditsEarned() {
        return creditsEarned;
    }

    public void setCreditsEarned(int creditsEarned) {
        this.creditsEarned = creditsEarned;
    }
}
