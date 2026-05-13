package com.example.registry.repository;

import com.example.registry.config.RegistryProperties;
import com.example.registry.model.InstanceStatus;
import com.example.registry.model.ServiceInstance;
import com.example.registry.model.Subscriber;
import org.springframework.stereotype.Repository;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class ServiceRegistryRepository {

    private final Map<String, Map<String, ServiceInstance>> registry = new ConcurrentHashMap<>();
    private final Map<String, List<Subscriber>> subscribers = new ConcurrentHashMap<>();
    private final RegistryProperties properties;

    public ServiceRegistryRepository(RegistryProperties properties) {
        this.properties = properties;
    }

    public ServiceInstance register(String serviceName, String ip, int port, 
                                    Map<String, String> metadata) {
        String instanceId = generateInstanceId(serviceName, ip, port);
        Map<String, ServiceInstance> instances = registry.computeIfAbsent(serviceName, 
            k -> new ConcurrentHashMap<>());

        ServiceInstance existing = instances.get(instanceId);
        if (existing != null && existing.isRecentlyDeregistered(
                properties.getHeartbeat().getRecoveryWindow())) {
            existing.setStatus(InstanceStatus.UP);
            existing.setLastHeartbeat(new java.util.Date().toInstant());
            existing.setHeartbeatFailures(0);
            existing.setDeregisterTime(null);
            existing.setMetadata(metadata != null ? metadata : existing.getMetadata());
            return existing;
        }

        ServiceInstance instance = ServiceInstance.builder()
            .instanceId(instanceId)
            .serviceName(serviceName)
            .ip(ip)
            .port(port)
            .metadata(metadata != null ? metadata : new HashMap<>())
            .status(InstanceStatus.UP)
            .leaseDuration(properties.getHeartbeat().getInterval() * 3)
            .lastHeartbeat(new java.util.Date().toInstant())
            .heartbeatFailures(0)
            .build();

        instances.put(instanceId, instance);
        return instance;
    }

    public Optional<ServiceInstance> findInstance(String serviceName, String instanceId) {
        Map<String, ServiceInstance> instances = registry.get(serviceName);
        return instances != null ? Optional.ofNullable(instances.get(instanceId)) : Optional.empty();
    }

    public List<ServiceInstance> findAllInstances(String serviceName) {
        Map<String, ServiceInstance> instances = registry.get(serviceName);
        return instances != null ? new ArrayList<>(instances.values()) : new ArrayList<>();
    }

    public List<ServiceInstance> findAllInstancesExcludingGoingDown(String serviceName) {
        return findAllInstances(serviceName).stream()
            .filter(i -> i.getStatus() != InstanceStatus.GOING_DOWN)
            .collect(Collectors.toList());
    }

    public List<ServiceInstance> findByMetadata(String serviceName, Map<String, String> metadataFilter) {
        return findAllInstancesExcludingGoingDown(serviceName).stream()
            .filter(instance -> matchesMetadata(instance, metadataFilter))
            .collect(Collectors.toList());
    }

    public void updateInstance(ServiceInstance instance) {
        Map<String, ServiceInstance> instances = registry.get(instance.getServiceName());
        if (instances != null) {
            instances.put(instance.getInstanceId(), instance);
        }
    }

    public void removeInstance(String serviceName, String instanceId) {
        Map<String, ServiceInstance> instances = registry.get(serviceName);
        if (instances != null) {
            instances.remove(instanceId);
            if (instances.isEmpty()) {
                registry.remove(serviceName);
            }
        }
    }

    public Map<String, Map<String, ServiceInstance>> getAllRegistry() {
        return registry;
    }

    public Subscriber subscribe(String serviceName, String callbackUrl) {
        String subscriberId = generateSubscriberId(serviceName, callbackUrl);
        List<Subscriber> serviceSubscribers = subscribers.computeIfAbsent(serviceName,
            k -> new ArrayList<>());

        Optional<Subscriber> existing = serviceSubscribers.stream()
            .filter(s -> s.getCallbackUrl().equals(callbackUrl))
            .findFirst();

        if (existing.isPresent()) {
            existing.get().setConsecutiveFailures(0);
            return existing.get();
        }

        Subscriber subscriber = Subscriber.builder()
            .subscriberId(subscriberId)
            .serviceName(serviceName)
            .callbackUrl(callbackUrl)
            .consecutiveFailures(0)
            .build();

        serviceSubscribers.add(subscriber);
        return subscriber;
    }

    public void unsubscribe(String serviceName, String callbackUrl) {
        List<Subscriber> serviceSubscribers = subscribers.get(serviceName);
        if (serviceSubscribers != null) {
            serviceSubscribers.removeIf(s -> s.getCallbackUrl().equals(callbackUrl));
            if (serviceSubscribers.isEmpty()) {
                subscribers.remove(serviceName);
            }
        }
    }

    public List<Subscriber> getSubscribers(String serviceName) {
        return subscribers.getOrDefault(serviceName, new ArrayList<>());
    }

    private String generateInstanceId(String serviceName, String ip, int port) {
        return serviceName + "-" + ip + ":" + port;
    }

    private String generateSubscriberId(String serviceName, String callbackUrl) {
        return "sub-" + serviceName + "-" + UUID.nameUUIDFromBytes(callbackUrl.getBytes());
    }

    private boolean matchesMetadata(ServiceInstance instance, Map<String, String> filter) {
        if (filter == null || filter.isEmpty()) {
            return true;
        }
        Map<String, String> metadata = instance.getMetadata();
        if (metadata == null) {
            return false;
        }
        return filter.entrySet().stream()
            .allMatch(entry -> entry.getValue().equals(metadata.get(entry.getKey())));
    }
}
