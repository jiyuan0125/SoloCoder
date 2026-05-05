package com.recruitment.server.repository;

import com.recruitment.common.dto.ApplicationDTO;
import com.recruitment.common.enums.ApplicationStatus;
import com.recruitment.common.enums.SourceChannel;
import com.recruitment.common.enums.Stage;
import org.springframework.stereotype.Repository;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class ApplicationRepository {
    private final Map<String, ApplicationDTO> applications = new ConcurrentHashMap<>();

    public ApplicationDTO save(ApplicationDTO application) {
        applications.put(application.getId(), application);
        return application;
    }

    public Optional<ApplicationDTO> findById(String id) {
        return Optional.ofNullable(applications.get(id));
    }

    public List<ApplicationDTO> findByCandidateId(String candidateId) {
        return applications.values().stream()
                .filter(a -> candidateId.equals(a.getCandidateId()))
                .collect(Collectors.toList());
    }

    public List<ApplicationDTO> findAll() {
        return new ArrayList<>(applications.values());
    }

    public List<ApplicationDTO> findByPosition(String position, Stage stage, SourceChannel sourceChannel) {
        return applications.values().stream()
                .filter(a -> position == null || position.isEmpty() || a.getPosition().contains(position))
                .filter(a -> stage == null || a.getCurrentStage() == stage)
                .collect(Collectors.toList());
    }

    public List<ApplicationDTO> findByCandidateIdExcluding(String candidateId, String excludedApplicationId) {
        return applications.values().stream()
                .filter(a -> candidateId.equals(a.getCandidateId()))
                .filter(a -> !excludedApplicationId.equals(a.getId()))
                .collect(Collectors.toList());
    }

    public List<ApplicationDTO> findByStatus(ApplicationStatus status) {
        return applications.values().stream()
                .filter(a -> a.getStatus() == status)
                .collect(Collectors.toList());
    }
}
