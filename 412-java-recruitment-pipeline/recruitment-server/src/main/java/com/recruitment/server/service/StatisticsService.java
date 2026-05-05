package com.recruitment.server.service;

import com.recruitment.common.dto.ApplicationDTO;
import com.recruitment.common.dto.StageStatisticsDTO;
import com.recruitment.common.dto.StageTransitionDTO;
import com.recruitment.common.enums.Stage;
import com.recruitment.common.response.ApiResponse;
import com.recruitment.common.response.StatisticsResponse;
import com.recruitment.server.repository.ApplicationRepository;
import com.recruitment.server.repository.CandidateRepository;
import com.recruitment.server.repository.EmployeeArchiveRepository;
import org.springframework.stereotype.Service;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.Duration;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class StatisticsService {
    private final ApplicationRepository applicationRepository;
    private final CandidateRepository candidateRepository;
    private final EmployeeArchiveRepository employeeArchiveRepository;

    public StatisticsService(ApplicationRepository applicationRepository,
                             CandidateRepository candidateRepository,
                             EmployeeArchiveRepository employeeArchiveRepository) {
        this.applicationRepository = applicationRepository;
        this.candidateRepository = candidateRepository;
        this.employeeArchiveRepository = employeeArchiveRepository;
    }

    public ApiResponse<StatisticsResponse> getStatistics() {
        StatisticsResponse response = new StatisticsResponse();
        response.setTotalCandidates(candidateRepository.findAll().size());
        response.setTotalOnboarded(employeeArchiveRepository.count());

        List<StageStatisticsDTO> stageStatistics = new ArrayList<>();
        Map<Stage, List<ApplicationDTO>> stageApplications = groupByStage();
        Map<Stage, Long> stageTotalTime = new HashMap<>();
        Map<Stage, Integer> stageTransitionCount = new HashMap<>();

        List<ApplicationDTO> allApplications = applicationRepository.findAll();
        for (ApplicationDTO app : allApplications) {
            calculateStageTime(app, stageTotalTime, stageTransitionCount);
        }

        for (Stage stage : Stage.values()) {
            StageStatisticsDTO stat = new StageStatisticsDTO();
            stat.setStage(stage);
            
            List<ApplicationDTO> appsInStage = stageApplications.getOrDefault(stage, new ArrayList<>());
            stat.setCount(appsInStage.size());

            int transitionCount = stageTransitionCount.getOrDefault(stage, 0);
            if (transitionCount > 0) {
                long totalTime = stageTotalTime.getOrDefault(stage, 0L);
                double avgDays = (double) totalTime / transitionCount / (1000.0 * 60 * 60 * 24);
                stat.setAvgDaysInStage(BigDecimal.valueOf(avgDays).setScale(2, RoundingMode.HALF_UP));
            }

            int prevStageCount = getPrevStageCount(stage, stageApplications);
            if (prevStageCount > 0) {
                BigDecimal conversionRate = BigDecimal.valueOf(appsInStage.size())
                        .divide(BigDecimal.valueOf(prevStageCount), 4, RoundingMode.HALF_UP)
                        .multiply(BigDecimal.valueOf(100));
                stat.setConversionRate(conversionRate);
            } else if (stage == Stage.SCREENING) {
                stat.setConversionRate(BigDecimal.valueOf(100));
            }

            stageStatistics.add(stat);
        }

        response.setStageStatistics(stageStatistics);
        return ApiResponse.success(response);
    }

    private Map<Stage, List<ApplicationDTO>> groupByStage() {
        Map<Stage, List<ApplicationDTO>> result = new HashMap<>();
        List<ApplicationDTO> applications = applicationRepository.findAll();
        for (ApplicationDTO app : applications) {
            result.computeIfAbsent(app.getCurrentStage(), k -> new ArrayList<>()).add(app);
        }
        return result;
    }

    private void calculateStageTime(ApplicationDTO app, Map<Stage, Long> stageTotalTime, Map<Stage, Integer> stageTransitionCount) {
        List<StageTransitionDTO> transitions = app.getStageTransitions();
        if (transitions.isEmpty()) {
            return;
        }

        for (int i = 0; i < transitions.size(); i++) {
            StageTransitionDTO transition = transitions.get(i);
            Stage stage = transition.getFromStage();
            LocalDateTime startTime = transition.getTransitionTime();
            LocalDateTime endTime;

            if (i + 1 < transitions.size()) {
                endTime = transitions.get(i + 1).getTransitionTime();
            } else {
                endTime = LocalDateTime.now();
            }

            long duration = Duration.between(startTime, endTime).toMillis();
            stageTotalTime.merge(stage, duration, Long::sum);
            stageTransitionCount.merge(stage, 1, Integer::sum);
        }
    }

    private int getPrevStageCount(Stage stage, Map<Stage, List<ApplicationDTO>> stageApplications) {
        Stage[] stages = Stage.values();
        int currentIndex = -1;
        for (int i = 0; i < stages.length; i++) {
            if (stages[i] == stage) {
                currentIndex = i;
                break;
            }
        }

        if (currentIndex <= 0) {
            return 0;
        }

        Stage prevStage = stages[currentIndex - 1];
        return stageApplications.getOrDefault(prevStage, new ArrayList<>()).size();
    }
}
