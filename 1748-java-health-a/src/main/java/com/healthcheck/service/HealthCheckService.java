package com.healthcheck.service;

import com.healthcheck.checker.Checker;
import com.healthcheck.checker.CheckerRegistry;
import com.healthcheck.model.*;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class HealthCheckService {

    private final CheckerRegistry checkerRegistry;
    private final StateMachineService stateMachineService;
    private final ServiceHealthCalculator healthCalculator;
    private final HistoryService historyService;
    private final WebhookNotifier webhookNotifier;
    private final ServiceConfigStore configStore;

    private final Map<String, ServiceState> serviceStates = new ConcurrentHashMap<>();

    public HealthCheckService(CheckerRegistry checkerRegistry,
                              StateMachineService stateMachineService,
                              ServiceHealthCalculator healthCalculator,
                              HistoryService historyService,
                              WebhookNotifier webhookNotifier,
                              ServiceConfigStore configStore) {
        this.checkerRegistry = checkerRegistry;
        this.stateMachineService = stateMachineService;
        this.healthCalculator = healthCalculator;
        this.historyService = historyService;
        this.webhookNotifier = webhookNotifier;
        this.configStore = configStore;
    }

    public void checkService(String serviceId) {
        Optional<ServiceConfig> configOpt = configStore.getServiceConfig(serviceId);
        if (configOpt.isEmpty()) {
            return;
        }

        ServiceConfig config = configOpt.get();
        if (!config.isEnabled()) {
            return;
        }

        ServiceState serviceState = serviceStates.computeIfAbsent(serviceId,
                id -> ServiceState.builder()
                        .serviceId(id)
                        .overallStatus(ServiceHealthStatus.UNKNOWN)
                        .previousStatus(ServiceHealthStatus.UNKNOWN)
                        .lastStateChangeTime(Instant.now())
                        .checkItemStates(new ConcurrentHashMap<>())
                        .build());

        for (CheckItemConfig itemConfig : config.getCheckItems()) {
            checkSingleItem(config, serviceState, itemConfig);
        }

        updateServiceOverallStatus(config, serviceState);
    }

    private void checkSingleItem(ServiceConfig serviceConfig,
                                 ServiceState serviceState,
                                 CheckItemConfig itemConfig) {
        Checker checker = checkerRegistry.getChecker(itemConfig);
        CheckResult rawResult = checker.execute(itemConfig, serviceConfig.getServiceId());

        historyService.addResult(rawResult);

        CheckItemState currentState = serviceState.getCheckItemStates().computeIfAbsent(
                itemConfig.getName(),
                name -> stateMachineService.createInitialState(
                        serviceConfig.getServiceId(), name));

        CheckItemState newState = stateMachineService.transitionState(currentState, rawResult);

        if (newState.getCurrentStatus() != currentState.getCurrentStatus()) {
            webhookNotifier.notifyCheckItemStatusChange(
                    serviceConfig.getServiceId(),
                    serviceConfig.getServiceName(),
                    itemConfig.getName(),
                    currentState.getCurrentStatus(),
                    newState.getCurrentStatus(),
                    serviceConfig.getWebhookUrl());
        }

        serviceState.getCheckItemStates().put(itemConfig.getName(), newState);
    }

    private void updateServiceOverallStatus(ServiceConfig config, ServiceState serviceState) {
        ServiceHealthStatus newOverallStatus = healthCalculator.calculate(
                serviceState.getCheckItemStates().values());

        ServiceHealthStatus oldStatus = serviceState.getOverallStatus();

        if (newOverallStatus != oldStatus) {
            serviceState.setPreviousStatus(oldStatus);
            serviceState.setOverallStatus(newOverallStatus);
            serviceState.setLastStateChangeTime(Instant.now());

            webhookNotifier.notifyServiceStatusChange(
                    config.getServiceId(),
                    config.getServiceName(),
                    oldStatus,
                    newOverallStatus,
                    config.getWebhookUrl());
        }
    }

    public Optional<ServiceState> getServiceState(String serviceId) {
        return Optional.ofNullable(serviceStates.get(serviceId));
    }

    public Map<String, ServiceState> getAllServiceStates() {
        return new ConcurrentHashMap<>(serviceStates);
    }

    public void initializeServiceState(ServiceConfig config) {
        if (!serviceStates.containsKey(config.getServiceId())) {
            ServiceState state = ServiceState.builder()
                    .serviceId(config.getServiceId())
                    .overallStatus(ServiceHealthStatus.UNKNOWN)
                    .previousStatus(ServiceHealthStatus.UNKNOWN)
                    .lastStateChangeTime(Instant.now())
                    .checkItemStates(new ConcurrentHashMap<>())
                    .build();
            
            for (CheckItemConfig itemConfig : config.getCheckItems()) {
                state.getCheckItemStates().put(
                        itemConfig.getName(),
                        stateMachineService.createInitialState(config.getServiceId(), itemConfig.getName()));
            }
            
            serviceStates.put(config.getServiceId(), state);
        }
    }
}
