package com.configcenter.service;

import com.configcenter.dto.ReleaseRequestDTO;
import com.configcenter.model.*;
import com.configcenter.repository.*;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.*;

@Slf4j
@Service
@RequiredArgsConstructor
public class ReleaseService {

    private final ReleaseRepository releaseRepository;
    private final ReleaseConfigChangeRepository configChangeRepository;
    private final ConfigItemRepository configItemRepository;
    private final ProjectRepository projectRepository;
    private final SubscriptionService subscriptionService;
    private final ValidationCallbackService validationCallbackService;
    private final ObjectMapper objectMapper;

    @Transactional
    public Release createRelease(Long projectId, ReleaseRequestDTO request) {
        Project project = projectRepository.findById(projectId)
                .orElseThrow(() -> new RuntimeException("Project not found: " + projectId));
        
        List<ConfigItem> pendingConfigs = configItemRepository.findByProjectIdAndEnvironmentAndStatus(
                projectId, 
                request.getEnvironment(), 
                ConfigItem.ReleaseStatus.PENDING
        );
        
        if (pendingConfigs.isEmpty()) {
            throw new RuntimeException("No pending configurations to release");
        }
        
        Release release = new Release();
        release.setProject(project);
        release.setEnvironment(request.getEnvironment());
        release.setType(request.getType());
        release.setStatus(Release.ReleaseStatus.PENDING);
        release.setCreatedBy(request.getCreatedBy());
        release.setStartTime(LocalDateTime.now());
        
        if (request.getType() == Release.ReleaseType.GRAY) {
            if (request.getGrayInstances() == null || request.getGrayInstances().isEmpty()) {
                throw new RuntimeException("Gray release requires at least one gray instance");
            }
            try {
                release.setGrayInstances(objectMapper.writeValueAsString(request.getGrayInstances()));
            } catch (JsonProcessingException e) {
                throw new RuntimeException("Failed to serialize gray instances", e);
            }
        }
        
        release = releaseRepository.save(release);
        
        for (ConfigItem config : pendingConfigs) {
            ReleaseConfigChange change = new ReleaseConfigChange();
            change.setRelease(release);
            change.setConfigItem(config);
            change.setConfigKey(config.getConfigKey());
            change.setOldValue(config.getCurrentValue());
            change.setNewValue(config.getPendingValue());
            configChangeRepository.save(change);
            
            config.setPendingReleaseId(release.getId());
            configItemRepository.save(config);
        }
        
        log.info("Release {} created for project {} environment {}", 
                release.getId(), projectId, request.getEnvironment());
        
        return release;
    }

    @Transactional
    public Release executeRelease(Long releaseId) {
        Release release = releaseRepository.findById(releaseId)
                .orElseThrow(() -> new RuntimeException("Release not found: " + releaseId));
        
        if (release.getStatus() != Release.ReleaseStatus.PENDING) {
            throw new RuntimeException("Release is not in PENDING state. Current state: " + release.getStatus());
        }
        
        release.setStatus(Release.ReleaseStatus.VALIDATING);
        releaseRepository.save(release);
        
        String callbackResult = validationCallbackService.validateRelease(release);
        release.setValidationCallbackResult(callbackResult);
        
        List<ReleaseConfigChange> configChanges = configChangeRepository.findByReleaseId(releaseId);
        
        if (release.getType() == Release.ReleaseType.GRAY) {
            release.setStatus(Release.ReleaseStatus.GRAY_IN_PROGRESS);
            releaseRepository.save(release);
            
            for (ReleaseConfigChange change : configChanges) {
                ConfigItem config = change.getConfigItem();
                config.setStatus(ConfigItem.ReleaseStatus.GRAY_RELEASE);
                config.setActiveReleaseId(release.getId());
                configItemRepository.save(config);
            }
            
            List<String> grayInstances = parseGrayInstances(release.getGrayInstances());
            
            subscriptionService.notifySubscribers(
                    release.getProject().getId(),
                    release.getEnvironment(),
                    Release.ReleaseType.GRAY,
                    grayInstances
            );
            
            log.info("Gray release {} started for project {} environment {} with {} gray instances",
                    releaseId, release.getProject().getId(), release.getEnvironment(), grayInstances.size());
            
        } else {
            release.setStatus(Release.ReleaseStatus.IN_PROGRESS);
            releaseRepository.save(release);
            
            for (ReleaseConfigChange change : configChanges) {
                ConfigItem config = change.getConfigItem();
                applyConfigChange(config, change);
            }
            
            release.setStatus(Release.ReleaseStatus.COMPLETED);
            release.setEndTime(LocalDateTime.now());
            releaseRepository.save(release);
            
            subscriptionService.notifySubscribers(
                    release.getProject().getId(),
                    release.getEnvironment(),
                    Release.ReleaseType.FULL,
                    null
            );
            
            log.info("Full release {} completed for project {} environment {}",
                    releaseId, release.getProject().getId(), release.getEnvironment());
        }
        
        return release;
    }

    @Transactional
    public Release fullReleaseGrayRelease(Long releaseId) {
        Release release = releaseRepository.findById(releaseId)
                .orElseThrow(() -> new RuntimeException("Release not found: " + releaseId));
        
        if (release.getStatus() != Release.ReleaseStatus.GRAY_IN_PROGRESS &&
            release.getStatus() != Release.ReleaseStatus.GRAY_COMPLETED) {
            throw new RuntimeException("Release is not in GRAY_IN_PROGRESS or GRAY_COMPLETED state");
        }
        
        List<ReleaseConfigChange> configChanges = configChangeRepository.findByReleaseId(releaseId);
        
        for (ReleaseConfigChange change : configChanges) {
            ConfigItem config = change.getConfigItem();
            applyConfigChange(config, change);
        }
        
        release.setStatus(Release.ReleaseStatus.COMPLETED);
        release.setEndTime(LocalDateTime.now());
        releaseRepository.save(release);
        
        subscriptionService.notifySubscribers(
                release.getProject().getId(),
                release.getEnvironment(),
                Release.ReleaseType.FULL,
                null
        );
        
        log.info("Gray release {} promoted to full release for project {} environment {}",
                releaseId, release.getProject().getId(), release.getEnvironment());
        
        return release;
    }

    @Transactional
    public Release rollbackRelease(Long releaseId) {
        Release release = releaseRepository.findById(releaseId)
                .orElseThrow(() -> new RuntimeException("Release not found: " + releaseId));
        
        if (release.getStatus() != Release.ReleaseStatus.GRAY_IN_PROGRESS &&
            release.getStatus() != Release.ReleaseStatus.COMPLETED) {
            throw new RuntimeException("Release cannot be rolled back from current state: " + release.getStatus());
        }
        
        List<ReleaseConfigChange> configChanges = configChangeRepository.findByReleaseId(releaseId);
        
        for (ReleaseConfigChange change : configChanges) {
            ConfigItem config = change.getConfigItem();
            config.setStatus(ConfigItem.ReleaseStatus.RELEASED);
            config.setPendingReleaseId(null);
            configItemRepository.save(config);
        }
        
        release.setStatus(Release.ReleaseStatus.ROLLED_BACK);
        release.setEndTime(LocalDateTime.now());
        releaseRepository.save(release);
        
        subscriptionService.notifySubscribers(
                release.getProject().getId(),
                release.getEnvironment(),
                Release.ReleaseType.FULL,
                null
        );
        
        log.info("Release {} rolled back for project {} environment {}",
                releaseId, release.getProject().getId(), release.getEnvironment());
        
        return release;
    }

    private void applyConfigChange(ConfigItem config, ReleaseConfigChange change) {
        config.setCurrentValue(change.getNewValue());
        config.setPendingValue(null);
        config.setStatus(ConfigItem.ReleaseStatus.RELEASED);
        config.setPendingReleaseId(null);
        config.setActiveReleaseId(change.getRelease().getId());
        configItemRepository.save(config);
    }

    private List<String> parseGrayInstances(String grayInstancesJson) {
        if (grayInstancesJson == null || grayInstancesJson.isEmpty()) {
            return Collections.emptyList();
        }
        try {
            return objectMapper.readValue(grayInstancesJson, new TypeReference<List<String>>() {});
        } catch (JsonProcessingException e) {
            log.warn("Failed to parse gray instances", e);
            return Collections.emptyList();
        }
    }

    public Release getRelease(Long releaseId) {
        return releaseRepository.findById(releaseId)
                .orElseThrow(() -> new RuntimeException("Release not found: " + releaseId));
    }

    public List<Release> getReleases(Long projectId, String environment) {
        return releaseRepository.findByProjectIdAndEnvironmentOrderByCreatedAtDesc(projectId, environment);
    }

    public List<Release> getActiveGrayReleases(Long projectId, String environment) {
        return releaseRepository.findByProjectIdAndEnvironmentAndStatusIn(
                projectId,
                environment,
                List.of(Release.ReleaseStatus.GRAY_IN_PROGRESS, Release.ReleaseStatus.GRAY_COMPLETED)
        );
    }
}
