package com.training.common.dto.response;

import java.util.List;

public class CourseEvaluationDTO {
    private String courseId;
    private String courseName;
    private int totalEnrollments;
    private int evaluationCount;
    private double averageRating;
    private List<StudentEvaluationDTO> evaluations;

    public CourseEvaluationDTO() {
    }

    public String getCourseId() {
        return courseId;
    }

    public void setCourseId(String courseId) {
        this.courseId = courseId;
    }

    public String getCourseName() {
        return courseName;
    }

    public void setCourseName(String courseName) {
        this.courseName = courseName;
    }

    public int getTotalEnrollments() {
        return totalEnrollments;
    }

    public void setTotalEnrollments(int totalEnrollments) {
        this.totalEnrollments = totalEnrollments;
    }

    public int getEvaluationCount() {
        return evaluationCount;
    }

    public void setEvaluationCount(int evaluationCount) {
        this.evaluationCount = evaluationCount;
    }

    public double getAverageRating() {
        return averageRating;
    }

    public void setAverageRating(double averageRating) {
        this.averageRating = averageRating;
    }

    public List<StudentEvaluationDTO> getEvaluations() {
        return evaluations;
    }

    public void setEvaluations(List<StudentEvaluationDTO> evaluations) {
        this.evaluations = evaluations;
    }
}
