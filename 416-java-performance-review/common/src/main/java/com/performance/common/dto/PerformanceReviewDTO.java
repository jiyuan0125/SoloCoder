package com.performance.common.dto;

import com.performance.common.enums.PerformanceGrade;
import com.performance.common.enums.ReviewStatus;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

public class PerformanceReviewDTO {
    private Long id;
    private Long employeeId;
    private String employeeName;
    private Long departmentId;
    private String departmentName;
    private Long cycleId;
    private String cycleName;
    private ReviewStatus status;
    
    private BigDecimal selfScore;
    private String selfComment;
    private LocalDateTime selfReviewTime;
    
    private Long managerId;
    private String managerName;
    private BigDecimal managerScore;
    private String managerComment;
    private LocalDateTime managerReviewTime;
    
    private List<OriginalDepartmentCommentDTO> originalDepartmentComments = new ArrayList<>();
    
    private BigDecimal finalScore;
    private PerformanceGrade finalGrade;
    private BigDecimal bonusCoefficient;
    
    private Integer ranking;
    private BigDecimal percentile;
    private Integer totalEmployees;
    
    private LocalDateTime completedTime;
    
    private Integer consecutiveCGrads;
    private Boolean pipEligible;
    private Boolean terminationEligible;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getEmployeeId() {
        return employeeId;
    }

    public void setEmployeeId(Long employeeId) {
        this.employeeId = employeeId;
    }

    public String getEmployeeName() {
        return employeeName;
    }

    public void setEmployeeName(String employeeName) {
        this.employeeName = employeeName;
    }

    public Long getDepartmentId() {
        return departmentId;
    }

    public void setDepartmentId(Long departmentId) {
        this.departmentId = departmentId;
    }

    public String getDepartmentName() {
        return departmentName;
    }

    public void setDepartmentName(String departmentName) {
        this.departmentName = departmentName;
    }

    public Long getCycleId() {
        return cycleId;
    }

    public void setCycleId(Long cycleId) {
        this.cycleId = cycleId;
    }

    public String getCycleName() {
        return cycleName;
    }

    public void setCycleName(String cycleName) {
        this.cycleName = cycleName;
    }

    public ReviewStatus getStatus() {
        return status;
    }

    public void setStatus(ReviewStatus status) {
        this.status = status;
    }

    public BigDecimal getSelfScore() {
        return selfScore;
    }

    public void setSelfScore(BigDecimal selfScore) {
        this.selfScore = selfScore;
    }

    public String getSelfComment() {
        return selfComment;
    }

    public void setSelfComment(String selfComment) {
        this.selfComment = selfComment;
    }

    public LocalDateTime getSelfReviewTime() {
        return selfReviewTime;
    }

    public void setSelfReviewTime(LocalDateTime selfReviewTime) {
        this.selfReviewTime = selfReviewTime;
    }

    public Long getManagerId() {
        return managerId;
    }

    public void setManagerId(Long managerId) {
        this.managerId = managerId;
    }

    public String getManagerName() {
        return managerName;
    }

    public void setManagerName(String managerName) {
        this.managerName = managerName;
    }

    public BigDecimal getManagerScore() {
        return managerScore;
    }

    public void setManagerScore(BigDecimal managerScore) {
        this.managerScore = managerScore;
    }

    public String getManagerComment() {
        return managerComment;
    }

    public void setManagerComment(String managerComment) {
        this.managerComment = managerComment;
    }

    public LocalDateTime getManagerReviewTime() {
        return managerReviewTime;
    }

    public void setManagerReviewTime(LocalDateTime managerReviewTime) {
        this.managerReviewTime = managerReviewTime;
    }

    public List<OriginalDepartmentCommentDTO> getOriginalDepartmentComments() {
        return originalDepartmentComments;
    }

    public void setOriginalDepartmentComments(List<OriginalDepartmentCommentDTO> originalDepartmentComments) {
        this.originalDepartmentComments = originalDepartmentComments;
    }

    public BigDecimal getFinalScore() {
        return finalScore;
    }

    public void setFinalScore(BigDecimal finalScore) {
        this.finalScore = finalScore;
    }

    public PerformanceGrade getFinalGrade() {
        return finalGrade;
    }

    public void setFinalGrade(PerformanceGrade finalGrade) {
        this.finalGrade = finalGrade;
    }

    public BigDecimal getBonusCoefficient() {
        return bonusCoefficient;
    }

    public void setBonusCoefficient(BigDecimal bonusCoefficient) {
        this.bonusCoefficient = bonusCoefficient;
    }

    public Integer getRanking() {
        return ranking;
    }

    public void setRanking(Integer ranking) {
        this.ranking = ranking;
    }

    public BigDecimal getPercentile() {
        return percentile;
    }

    public void setPercentile(BigDecimal percentile) {
        this.percentile = percentile;
    }

    public Integer getTotalEmployees() {
        return totalEmployees;
    }

    public void setTotalEmployees(Integer totalEmployees) {
        this.totalEmployees = totalEmployees;
    }

    public LocalDateTime getCompletedTime() {
        return completedTime;
    }

    public void setCompletedTime(LocalDateTime completedTime) {
        this.completedTime = completedTime;
    }

    public Integer getConsecutiveCGrads() {
        return consecutiveCGrads;
    }

    public void setConsecutiveCGrads(Integer consecutiveCGrads) {
        this.consecutiveCGrads = consecutiveCGrads;
    }

    public Boolean getPipEligible() {
        return pipEligible;
    }

    public void setPipEligible(Boolean pipEligible) {
        this.pipEligible = pipEligible;
    }

    public Boolean getTerminationEligible() {
        return terminationEligible;
    }

    public void setTerminationEligible(Boolean terminationEligible) {
        this.terminationEligible = terminationEligible;
    }
}