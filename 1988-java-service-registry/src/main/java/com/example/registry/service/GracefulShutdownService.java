package com.example.registry.service;

import com.example.registry.config.RegistryProperties;
import com.example.registry.model.InstanceStatus;
import com.example.registry.model.ServiceInstance;
import com.example.registry.repository.ServiceRegistryRepository;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.Map;
import java.util.Timer;
import java.util.TimerTask;
import java.util.concurrent.ConcurrentHashMap;

@Slf4j
@Service
public class GracefulShutdownService {

    private final ServiceRegistryRepository repository;
    private final RegistryProperties properties;
    private final Timer timer = new Timer("graceful-shutdown", true);
    private final Map<String, TimerTask> scheduledTasks = new ConcurrentHashMap<>();

    public GracefulShutdownService(ServiceRegistryRepository repository,
                                   RegistryProperties properties) {
        this.repository = repository;
        this.properties = properties;
    }

    public boolean initiateGracefulShutdown(String serviceName, String instanceId) {
        ServiceInstance instance = repository.findInstance(serviceName, instanceId)
            .orElse(null);
        
        if (instance == null || instance.getStatus() == InstanceStatus.DEREGISTERED) {
            return false;
        }

        if (instance.getStatus() == InstanceStatus.GOING_DOWN) {
            return true;
        }

        log.info("Initiating graceful shutdown for instance {}", instanceId);
        instance.setStatus(InstanceStatus.GOING_DOWN);
        repository.updateInstance(instance);

        TimerTask task = new TimerTask() {
            @Override
            public void run() {
                log.info("Completing graceful shutdown for instance {}", instanceId);
                instance.setStatus(InstanceStatus.DEREGISTERED);
                instance.setDeregisterTime(java.time.Instant.now());
                repository.updateInstance(instance);
                scheduledTasks.remove(instanceId);
            }
        };

        scheduledTasks.put(instanceId, task);
        timer.schedule(task, properties.getGracefulShutdown().getWaitTime());
        return true;
    }

    public boolean cancelGracefulShutdown(String serviceName, String instanceId) {
        TimerTask task = scheduledTasks.remove(instanceId);
        if (task != null) {
            task.cancel();
            ServiceInstance instance = repository.findInstance(serviceName, instanceId)
                .orElse(null);
            if (instance != null && instance.getStatus() == InstanceStatus.GOING_DOWN) {
                instance.setStatus(InstanceStatus.UP);
                instance.setHeartbeatFailures(0);
                repository.updateInstance(instance);
                log.info("Cancelled graceful shutdown for instance {}", instanceId);
                return true;
            }
        }
        return false;
    }
}
