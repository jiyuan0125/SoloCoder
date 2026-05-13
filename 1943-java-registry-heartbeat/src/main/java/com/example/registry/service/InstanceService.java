package com.example.registry.service;

import com.example.registry.dto.RegisterRequest;
import com.example.registry.model.ChangeEvent;
import com.example.registry.model.ServiceInstance;
import com.example.registry.store.InstanceStore;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.Set;
import java.util.UUID;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class InstanceService {
    private final InstanceStore instanceStore;
    private final NotificationService notificationService;

    public ServiceInstance register(String serviceName, RegisterRequest request) {
        String instanceId = UUID.randomUUID().toString();
        Instant now = Instant.now();

        ServiceInstance instance = ServiceInstance.builder()
                .id(instanceId)
                .serviceName(serviceName)
                .ip(request.getIp())
                .port(request.getPort())
                .weight(request.getWeight() != null ? request.getWeight() : 1)
                .heartbeatTtlSeconds(request.getHeartbeatTtlSeconds())
                .healthStatus(ServiceInstance.HealthStatus.HEALTHY)
                .lastHeartbeat(now)
                .registerTime(now)
                .missedHeartbeats(0)
                .build();

        ServiceInstance saved = instanceStore.save(instance);
        log.info("Instance registered: {} - {}:{}", saved.getId(), saved.getIp(), saved.getPort());

        notificationService.notifySubscribers(serviceName, 
            ChangeEvent.builder().type(ChangeEvent.ChangeType.REGISTER).instance(saved).build());

        return saved;
    }

    public Optional<ServiceInstance> heartbeat(String instanceId) {
        return instanceStore.findById(instanceId).map(instance -> {
            instance.setLastHeartbeat(Instant.now());
            instance.setMissedHeartbeats(0);
            ServiceInstance.HealthStatus oldStatus = instance.getHealthStatus();
            if (oldStatus == ServiceInstance.HealthStatus.UNHEALTHY) {
                instance.setHealthStatus(ServiceInstance.HealthStatus.HEALTHY);
                instanceStore.save(instance);
                notificationService.notifySubscribers(instance.getServiceName(),
                    ChangeEvent.builder().type(ChangeEvent.ChangeType.STATUS_CHANGE).instance(instance).build());
            } else {
                instanceStore.save(instance);
            }
            log.debug("Heartbeat received for instance: {}", instanceId);
            return instance;
        });
    }

    public List<ServiceInstance> getByServiceName(String serviceName, boolean healthyOnly) {
        List<ServiceInstance> instances = instanceStore.findByServiceName(serviceName);
        if (healthyOnly) {
            return instances.stream()
                    .filter(i -> i.getHealthStatus() == ServiceInstance.HealthStatus.HEALTHY)
                    .collect(Collectors.toList());
        }
        return instances;
    }

    public Set<String> getAllServiceNames() {
        return instanceStore.getAllServiceNames();
    }

    public int countByServiceName(String serviceName) {
        return instanceStore.countByServiceName(serviceName);
    }

    public Optional<ServiceInstance> findById(String instanceId) {
        return instanceStore.findById(instanceId);
    }

    public void deregister(String instanceId) {
        instanceStore.findById(instanceId).ifPresent(instance -> {
            instanceStore.remove(instanceId);
            log.info("Instance deregistered: {}", instanceId);
            notificationService.notifySubscribers(instance.getServiceName(),
                ChangeEvent.builder().type(ChangeEvent.ChangeType.DEREGISTER).instance(instance).build());
        });
    }

    public void markUnhealthyAndDeregister(ServiceInstance instance) {
        instance.setHealthStatus(ServiceInstance.HealthStatus.UNHEALTHY);
        instanceStore.save(instance);
        log.info("Instance marked as unhealthy: {}", instance.getId());

        notificationService.notifySubscribers(instance.getServiceName(),
            ChangeEvent.builder().type(ChangeEvent.ChangeType.STATUS_CHANGE).instance(instance).build());

        instanceStore.remove(instance.getId());
        log.info("Instance automatically deregistered: {}", instance.getId());

        notificationService.notifySubscribers(instance.getServiceName(),
            ChangeEvent.builder().type(ChangeEvent.ChangeType.DEREGISTER).instance(instance).build());
    }
}
