package com.training.common.dto.request;

import com.training.common.enums.CourseType;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;

public class CreateCourseRequest {
    private String name;
    private String description;
    private String instructorId;
    private int durationMinutes;
    private int maxCapacity;
    private LocalDateTime startTime;
    private LocalDateTime endTime;
    private CourseType courseType;
    private List<String> requiredDepartments;
    private int credits;
    private BigDecimal fee;

    public CreateCourseRequest() {
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

    public String getInstructorId() {
        return instructorId;
    }

    public void setInstructorId(String instructorId) {
        this.instructorId = instructorId;
    }

    public int getDurationMinutes() {
        return durationMinutes;
    }

    public void setDurationMinutes(int durationMinutes) {
        this.durationMinutes = durationMinutes;
    }

    public int getMaxCapacity() {
        return maxCapacity;
    }

    public void setMaxCapacity(int maxCapacity) {
        this.maxCapacity = maxCapacity;
    }

    public LocalDateTime getStartTime() {
        return startTime;
    }

    public void setStartTime(LocalDateTime startTime) {
        this.startTime = startTime;
    }

    public LocalDateTime getEndTime() {
        return endTime;
    }

    public void setEndTime(LocalDateTime endTime) {
        this.endTime = endTime;
    }

    public CourseType getCourseType() {
        return courseType;
    }

    public void setCourseType(CourseType courseType) {
        this.courseType = courseType;
    }

    public List<String> getRequiredDepartments() {
        return requiredDepartments;
    }

    public void setRequiredDepartments(List<String> requiredDepartments) {
        this.requiredDepartments = requiredDepartments;
    }

    public int getCredits() {
        return credits;
    }

    public void setCredits(int credits) {
        this.credits = credits;
    }

    public BigDecimal getFee() {
        return fee;
    }

    public void setFee(BigDecimal fee) {
        this.fee = fee;
    }
}
