package com.training.common.dto.request;

import com.training.common.enums.CourseStatus;
import com.training.common.enums.CourseType;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;

public class UpdateCourseRequest {
    private String name;
    private String description;
    private String instructorId;
    private Integer durationMinutes;
    private Integer maxCapacity;
    private LocalDateTime startTime;
    private LocalDateTime endTime;
    private CourseType courseType;
    private List<String> requiredDepartments;
    private CourseStatus status;
    private Integer credits;
    private BigDecimal fee;

    public UpdateCourseRequest() {
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

    public Integer getDurationMinutes() {
        return durationMinutes;
    }

    public void setDurationMinutes(Integer durationMinutes) {
        this.durationMinutes = durationMinutes;
    }

    public Integer getMaxCapacity() {
        return maxCapacity;
    }

    public void setMaxCapacity(Integer maxCapacity) {
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

    public CourseStatus getStatus() {
        return status;
    }

    public void setStatus(CourseStatus status) {
        this.status = status;
    }

    public Integer getCredits() {
        return credits;
    }

    public void setCredits(Integer credits) {
        this.credits = credits;
    }

    public BigDecimal getFee() {
        return fee;
    }

    public void setFee(BigDecimal fee) {
        this.fee = fee;
    }
}
