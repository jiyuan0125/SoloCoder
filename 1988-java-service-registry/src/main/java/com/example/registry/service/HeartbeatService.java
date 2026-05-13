package com.example.registry.service;

import com.example.registry.config.RegistryProperties;
import com.example.registry.model.InstanceStatus;
import com.example.registry.model.ServiceInstance;
import com.example.registry.repository.ServiceRegistryRepository;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.time.Duration;
import java.time.Instant;
import java.util.Map;
import java.util.Optional;

@Slf4j
@Service
public class HeartbeatService {

    private final ServiceRegistryRepository repository;
    private final RegistryProperties properties;

    public HeartbeatService(ServiceRegistryRepository repository,
                            RegistryProperties properties) {
        this.repository = repository;
        this.properties = properties;
    }

    public boolean heartbeat(String serviceName, String instanceId) {
        Optional<ServiceInstance> opt = repository.findInstance(serviceName, instanceId);
        if (!opt.isPresent()) {
            return false;
        }

        ServiceInstance instance = opt.get();
        if (instance.getStatus() == InstanceStatus.DEREGISTERED) {
            return false;
        }

        instance.setLastHeartbeat(Instant.now());
        instance.setHeartbeatFailures(0);
        instance.setStatus(InstanceStatus.UP);
        repository.updateInstance(instance);
        return true;
    }

    @Scheduled(fixedRateString = "${registry.heartbeat.interval:10000}")
    public void checkHeartbeats() {
        for (Map.Entry<String, Map<String, ServiceInstance>> entry : repository.getAllRegistry().entrySet()) {
            for (ServiceInstance instance : entry.getValue().values()) {
                checkInstanceHeartbeat(instance);
            }
        }
    }

    private void checkInstanceHeartbeat(ServiceInstance instance) {
        if (instance.getStatus() == InstanceStatus.GOING_DOWN ||
            instance.getStatus() == InstanceStatus.DEREGISTERED) {
            return;
        }

        Instant lastHeartbeat = instance.getLastHeartbeat();
        if (lastHeartbeat == null) {
            return;
        }

        long elapsed = Duration.between(lastHeartbeat, Instant.now()).toMillis();
        if (elapsed > properties.getHeartbeat().getInterval()) {
            instance.setHeartbeatFailures(instance.getHeartbeatFailures() + 1);
            log.debug("Instance {} heartbeat failure count: {}", 
                instance.getInstanceId(), instance.getHeartbeatFailures());

            if (instance.getHeartbeatFailures() >= properties.getHeartbeat().getUnhealthyThreshold() &&
                instance.getStatus() == InstanceStatus.UP) {
                log.info("Marking instance {} as UNHEALTHY", instance.getInstanceId());
                instance.setStatus(InstanceStatus.UNHEALTHY);
            }

            if (instance.getHeartbeatFailures() >= properties.getHeartbeat().getDeregisterThreshold()) {
                log.info("Deregistering instance {} due to heartbeats", instance.getInstanceId());
                instance.setStatus(InstanceStatus.DEREGISTERED);
                instance.setDeregisterTime(Instant.now());
            }

            repository.updateInstance(instance);
        }
    }
}
