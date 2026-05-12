package com.configcenter.service;

import com.configcenter.dto.ConfigChangeEvent;
import com.configcenter.model.ClientSubscription;
import com.configcenter.model.ConfigItem;
import com.configcenter.model.Release;
import com.configcenter.repository.ClientSubscriptionRepository;
import com.configcenter.repository.ConfigItemRepository;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

@Slf4j
@Service
@RequiredArgsConstructor
public class SubscriptionService {

    private final ClientSubscriptionRepository clientSubscriptionRepository;
    private final ConfigItemRepository configItemRepository;
    private final ObjectMapper objectMapper;

    @Value("${config.long-polling.timeout:30000}")
    private long longPollingTimeout;

    private final Map<String, CountDownLatch> waiters = new ConcurrentHashMap<>();

    @Transactional
    public ClientSubscription subscribe(String instanceId, Long projectId, String environment, String lastKnownVersion) {
        ClientSubscription subscription = clientSubscriptionRepository
                .findByInstanceIdAndProjectIdAndEnvironment(instanceId, projectId, environment)
                .orElseGet(() -> {
                    ClientSubscription newSub = new ClientSubscription();
                    newSub.setInstanceId(instanceId);
                    newSub.setProjectId(projectId);
                    newSub.setEnvironment(environment);
                    return newSub;
                });

        subscription.setLastKnownVersion(lastKnownVersion);
        subscription.setLastHeartbeat(LocalDateTime.now());
        
        return clientSubscriptionRepository.save(subscription);
    }

    @Transactional
    public void heartbeat(String instanceId, Long projectId, String environment) {
        clientSubscriptionRepository
                .findByInstanceIdAndProjectIdAndEnvironment(instanceId, projectId, environment)
                .ifPresent(sub -> {
                    sub.setLastHeartbeat(LocalDateTime.now());
                    clientSubscriptionRepository.save(sub);
                });
    }

    public ConfigChangeEvent waitForChanges(String instanceId, Long projectId, String environment, String currentVersion) {
        String waiterKey = instanceId + "-" + projectId + "-" + environment;
        
        List<ConfigItem> currentConfigs = configItemRepository
                .findByProjectIdAndEnvironment(projectId, environment);
        String latestVersion = calculateVersion(currentConfigs);
        
        if (currentVersion == null || !currentVersion.equals(latestVersion)) {
            return buildChangeEvent(projectId, environment, currentConfigs);
        }
        
        CountDownLatch latch = new CountDownLatch(1);
        waiters.put(waiterKey, latch);
        
        try {
            boolean signaled = latch.await(longPollingTimeout, TimeUnit.MILLISECONDS);
            if (signaled) {
                List<ConfigItem> updatedConfigs = configItemRepository
                        .findByProjectIdAndEnvironment(projectId, environment);
                String newVersion = calculateVersion(updatedConfigs);
                
                if (!newVersion.equals(currentVersion)) {
                    return buildChangeEvent(projectId, environment, updatedConfigs);
                }
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            log.warn("Long polling interrupted for {}", waiterKey);
        } finally {
            waiters.remove(waiterKey);
        }
        
        return null;
    }

    public void notifySubscribers(Long projectId, String environment, Release.ReleaseType releaseType, List<String> grayInstances) {
        List<ClientSubscription> subscriptions = clientSubscriptionRepository
                .findByProjectIdAndEnvironment(projectId, environment);
        
        for (ClientSubscription sub : subscriptions) {
            boolean shouldNotify = releaseType == Release.ReleaseType.FULL ||
                    (grayInstances != null && grayInstances.contains(sub.getInstanceId()));
            
            if (shouldNotify) {
                String waiterKey = sub.getInstanceId() + "-" + projectId + "-" + environment;
                CountDownLatch latch = waiters.get(waiterKey);
                if (latch != null) {
                    latch.countDown();
                    log.info("Notified subscriber {} for project {} environment {}", 
                            sub.getInstanceId(), projectId, environment);
                }
            }
        }
    }

    private String calculateVersion(List<ConfigItem> configs) {
        Map<String, String> releasedConfigs = new HashMap<>();
        for (ConfigItem config : configs) {
            if (config.getStatus() == ConfigItem.ReleaseStatus.RELEASED) {
                releasedConfigs.put(config.getConfigKey(), config.getCurrentValue());
            }
        }
        
        try {
            return Integer.toHexString(objectMapper.writeValueAsString(releasedConfigs).hashCode());
        } catch (JsonProcessingException e) {
            return String.valueOf(System.currentTimeMillis());
        }
    }

    private ConfigChangeEvent buildChangeEvent(Long projectId, String environment, List<ConfigItem> configs) {
        Map<String, String> changes = new HashMap<>();
        for (ConfigItem config : configs) {
            if (config.getStatus() == ConfigItem.ReleaseStatus.RELEASED) {
                changes.put(config.getConfigKey(), config.getCurrentValue());
            }
        }
        
        ConfigChangeEvent event = new ConfigChangeEvent();
        event.setProjectId(projectId.toString());
        event.setEnvironment(environment);
        event.setVersion(calculateVersion(configs));
        event.setChanges(changes);
        event.setTimestamp(LocalDateTime.now());
        event.setGrayRelease(false);
        return event;
    }

    public List<ClientSubscription> getSubscribers(Long projectId, String environment) {
        return clientSubscriptionRepository.findByProjectIdAndEnvironment(projectId, environment);
    }
}
