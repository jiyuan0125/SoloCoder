package com.recruitment.common.dto;

import com.recruitment.common.enums.ApplicationStatus;
import com.recruitment.common.enums.Stage;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

public class ApplicationDTO {
    private String id;
    private String candidateId;
    private String position;
    private Stage currentStage;
    private ApplicationStatus status;
    private String abandonReason;
    private boolean notRecommended;
    private List<InterviewRoundDTO> interviewRounds;
    private OfferDTO offer;
    private List<StageTransitionDTO> stageTransitions;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;

    public ApplicationDTO() {
        this.interviewRounds = new ArrayList<>();
        this.stageTransitions = new ArrayList<>();
        this.status = ApplicationStatus.IN_PROGRESS;
        this.currentStage = Stage.SCREENING;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getCandidateId() {
        return candidateId;
    }

    public void setCandidateId(String candidateId) {
        this.candidateId = candidateId;
    }

    public String getPosition() {
        return position;
    }

    public void setPosition(String position) {
        this.position = position;
    }

    public Stage getCurrentStage() {
        return currentStage;
    }

    public void setCurrentStage(Stage currentStage) {
        this.currentStage = currentStage;
    }

    public ApplicationStatus getStatus() {
        return status;
    }

    public void setStatus(ApplicationStatus status) {
        this.status = status;
    }

    public String getAbandonReason() {
        return abandonReason;
    }

    public void setAbandonReason(String abandonReason) {
        this.abandonReason = abandonReason;
    }

    public boolean isNotRecommended() {
        return notRecommended;
    }

    public void setNotRecommended(boolean notRecommended) {
        this.notRecommended = notRecommended;
    }

    public List<InterviewRoundDTO> getInterviewRounds() {
        return interviewRounds;
    }

    public void setInterviewRounds(List<InterviewRoundDTO> interviewRounds) {
        this.interviewRounds = interviewRounds;
    }

    public OfferDTO getOffer() {
        return offer;
    }

    public void setOffer(OfferDTO offer) {
        this.offer = offer;
    }

    public List<StageTransitionDTO> getStageTransitions() {
        return stageTransitions;
    }

    public void setStageTransitions(List<StageTransitionDTO> stageTransitions) {
        this.stageTransitions = stageTransitions;
    }

    public LocalDateTime getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(LocalDateTime createdAt) {
        this.createdAt = createdAt;
    }

    public LocalDateTime getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(LocalDateTime updatedAt) {
        this.updatedAt = updatedAt;
    }

    public BigDecimal getAverageScore() {
        if (interviewRounds == null || interviewRounds.isEmpty()) {
            return BigDecimal.ZERO;
        }
        int totalScore = 0;
        int validCount = 0;
        for (InterviewRoundDTO round : interviewRounds) {
            if (round.getScore() != null) {
                totalScore += round.getScore();
                validCount++;
            }
        }
        if (validCount == 0) {
            return BigDecimal.ZERO;
        }
        return BigDecimal.valueOf(totalScore).divide(BigDecimal.valueOf(validCount), 2, RoundingMode.HALF_UP);
    }
}
