package com.training.common.dto.response;

import java.util.List;

public class InstructorEvaluationSummaryDTO {
    private String instructorId;
    private String instructorName;
    private int totalCourses;
    private int totalEvaluations;
    private double averageRating;
    private List<CourseEvaluationDTO> courseEvaluations;

    public InstructorEvaluationSummaryDTO() {
    }

    public String getInstructorId() {
        return instructorId;
    }

    public void setInstructorId(String instructorId) {
        this.instructorId = instructorId;
    }

    public String getInstructorName() {
        return instructorName;
    }

    public void setInstructorName(String instructorName) {
        this.instructorName = instructorName;
    }

    public int getTotalCourses() {
        return totalCourses;
    }

    public void setTotalCourses(int totalCourses) {
        this.totalCourses = totalCourses;
    }

    public int getTotalEvaluations() {
        return totalEvaluations;
    }

    public void setTotalEvaluations(int totalEvaluations) {
        this.totalEvaluations = totalEvaluations;
    }

    public double getAverageRating() {
        return averageRating;
    }

    public void setAverageRating(double averageRating) {
        this.averageRating = averageRating;
    }

    public List<CourseEvaluationDTO> getCourseEvaluations() {
        return courseEvaluations;
    }

    public void setCourseEvaluations(List<CourseEvaluationDTO> courseEvaluations) {
        this.courseEvaluations = courseEvaluations;
    }
}
