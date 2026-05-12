package com.example.serviceregistry.service;

import com.example.serviceregistry.dto.RegisterRequest;
import com.example.serviceregistry.event.ServiceChangedEvent;
import com.example.serviceregistry.model.ServiceInstance;
import com.example.serviceregistry.model.Subscriber;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.context.ApplicationEventPublisher;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.time.temporal.ChronoUnit;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class ServiceRegistry {

    private final ApplicationEventPublisher eventPublisher;

    private final Map<String, Map<String, ServiceInstance>> registry = new ConcurrentHashMap<>();
    private final Map<String, Set<Subscriber>> subscribers = new ConcurrentHashMap<>();

    public ServiceInstance register(RegisterRequest request) {
        ServiceInstance instance = ServiceInstance.builder()
                .instanceId(request.getInstanceId())
                .serviceName(request.getServiceName())
                .ip(request.getIp())
                .port(request.getPort())
                .tags(request.getTags() != null ? request.getTags() : Map.of())
                .weight(request.getWeight() != null ? request.getWeight() : 1)
                .heartbeatInterval(request.getHeartbeatInterval() != null ? request.getHeartbeatInterval() : 30)
                .registeredAt(Instant.now())
                .lastHeartbeatAt(Instant.now())
                .healthy(true)
                .build();

        registry.computeIfAbsent(request.getServiceName(), k -> new ConcurrentHashMap<>())
                .put(request.getInstanceId(), instance);

        log.info("Service registered: {} - {}:{}", request.getServiceName(), request.getIp(), request.getPort());

        eventPublisher.publishEvent(new ServiceChangedEvent(this, request.getServiceName(), getHealthyInstances(request.getServiceName())));

        return instance;
    }

    public Optional<ServiceInstance> heartbeat(String serviceName, String instanceId) {
        Map<String, ServiceInstance> instances = registry.get(serviceName);
        if (instances == null) {
            return Optional.empty();
        }

        ServiceInstance instance = instances.get(instanceId);
        if (instance == null) {
            return Optional.empty();
        }

        instance.setLastHeartbeatAt(Instant.now());
        instance.setHealthy(true);
        return Optional.of(instance);
    }

    public void deregister(String serviceName, String instanceId) {
        Map<String, ServiceInstance> instances = registry.get(serviceName);
        if (instances != null) {
            instances.remove(instanceId);
            log.info("Service deregistered: {} - {}", serviceName, instanceId);
            eventPublisher.publishEvent(new ServiceChangedEvent(this, serviceName, getHealthyInstances(serviceName)));
        }
    }

    public List<ServiceInstance> getAllInstances() {
        return registry.values().stream()
                .flatMap(map -> map.values().stream())
                .collect(Collectors.toList());
    }

    public List<ServiceInstance> getAllInstancesWithTags(Map<String, String> tags) {
        if (tags == null || tags.isEmpty()) {
            return getAllInstances();
        }
        return getAllInstances().stream()
                .filter(instance -> {
                    Map<String, String> instanceTags = instance.getTags();
                    return tags.entrySet().stream()
                            .allMatch(entry -> entry.getValue().equals(instanceTags.get(entry.getKey())));
                })
                .collect(Collectors.toList());
    }

    public Optional<ServiceInstance> getInstance(String serviceName, String instanceId) {
        Map<String, ServiceInstance> instances = registry.get(serviceName);
        if (instances == null) {
            return Optional.empty();
        }
        return Optional.ofNullable(instances.get(instanceId));
    }

    public List<ServiceInstance> getHealthyInstances(String serviceName) {
        Map<String, ServiceInstance> instances = registry.get(serviceName);
        if (instances == null) {
            return List.of();
        }
        return instances.values().stream()
                .filter(ServiceInstance::isHealthy)
                .filter(instance -> instance.getWeight() > 0)
                .collect(Collectors.toList());
    }

    public void subscribe(Subscriber subscriber) {
        subscribers.computeIfAbsent(subscriber.getServiceName(), k -> Collections.newSetFromMap(new ConcurrentHashMap<>()))
                .add(subscriber);
        log.info("Subscriber registered: {} -> {}", subscriber.getSubscriberId(), subscriber.getCallbackUrl());
    }

    public void unsubscribe(String subscriberId, String serviceName) {
        Set<Subscriber> serviceSubscribers = subscribers.get(serviceName);
        if (serviceSubscribers != null) {
            serviceSubscribers.removeIf(s -> s.getSubscriberId().equals(subscriberId));
        }
    }

    public Set<Subscriber> getSubscribers(String serviceName) {
        return subscribers.getOrDefault(serviceName, Set.of());
    }

    @Scheduled(fixedDelay = 5000)
    public void healthCheck() {
        Instant now = Instant.now();
        List<String> servicesToNotify = new ArrayList<>();

        registry.forEach((serviceName, instances) -> {
            List<String> toRemove = new ArrayList<>();

            instances.forEach((instanceId, instance) -> {
                long secondsSinceLastHeartbeat = ChronoUnit.SECONDS.between(instance.getLastHeartbeatAt(), now);
                long timeout = instance.getHeartbeatInterval() * 3L;

                if (secondsSinceLastHeartbeat > timeout) {
                    log.warn("Service instance unhealthy: {} - {} - {}s since last heartbeat",
                            serviceName, instanceId, secondsSinceLastHeartbeat);
                    if (instance.isHealthy()) {
                        instance.setHealthy(false);
                        if (!servicesToNotify.contains(serviceName)) {
                            servicesToNotify.add(serviceName);
                        }
                    }

                    if (secondsSinceLastHeartbeat > timeout * 3) {
                        toRemove.add(instanceId);
                    }
                }
            });

            if (!toRemove.isEmpty()) {
                toRemove.forEach(instances::remove);
                log.info("Removed {} unhealthy instances from service: {}", toRemove.size(), serviceName);
                if (!servicesToNotify.contains(serviceName)) {
                    servicesToNotify.add(serviceName);
                }
            }
        });

        servicesToNotify.forEach(serviceName ->
                eventPublisher.publishEvent(new ServiceChangedEvent(this, serviceName, getHealthyInstances(serviceName))));
    }
}
