package com.recruitment.common.response;

import com.recruitment.common.dto.StageStatisticsDTO;

import java.util.List;

public class StatisticsResponse {
    private List<StageStatisticsDTO> stageStatistics;
    private Integer totalCandidates;
    private Integer totalOnboarded;

    public StatisticsResponse() {
    }

    public List<StageStatisticsDTO> getStageStatistics() {
        return stageStatistics;
    }

    public void setStageStatistics(List<StageStatisticsDTO> stageStatistics) {
        this.stageStatistics = stageStatistics;
    }

    public Integer getTotalCandidates() {
        return totalCandidates;
    }

    public void setTotalCandidates(Integer totalCandidates) {
        this.totalCandidates = totalCandidates;
    }

    public Integer getTotalOnboarded() {
        return totalOnboarded;
    }

    public void setTotalOnboarded(Integer totalOnboarded) {
        this.totalOnboarded = totalOnboarded;
    }
}
