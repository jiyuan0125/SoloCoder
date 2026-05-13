package com.example.registry.scheduler;

import com.example.registry.config.RegistryProperties;
import com.example.registry.model.ServiceInstance;
import com.example.registry.service.InstanceService;
import com.example.registry.store.InstanceStore;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

@Slf4j
@Component
@RequiredArgsConstructor
public class HeartbeatChecker {
    private final InstanceStore instanceStore;
    private final InstanceService instanceService;
    private final RegistryProperties registryProperties;

    @Scheduled(fixedDelayString = "${registry.heartbeat-check-interval-ms:5000}")
    public void checkHeartbeats() {
        List<ServiceInstance> allInstances = instanceStore.findAll();
        if (allInstances.isEmpty()) {
            return;
        }

        Instant now = Instant.now();
        int missedThreshold = registryProperties.getHeartbeatMissedThreshold();
        List<String> instancesToRemove = new ArrayList<>();

        for (ServiceInstance instance : allInstances) {
            long ttlMillis = (long) instance.getHeartbeatTtlSeconds() * 1000;
            long elapsed = now.toEpochMilli() - instance.getLastHeartbeat().toEpochMilli();

            if (elapsed > ttlMillis) {
                int missed = instance.getMissedHeartbeats() + 1;
                instance.setMissedHeartbeats(missed);
                log.debug("Instance {} missed heartbeat {}/{}", instance.getId(), missed, missedThreshold);

                if (missed >= missedThreshold) {
                    log.warn("Instance {} missed {} heartbeats, marking unhealthy and deregistering",
                            instance.getId(), missed);
                    instancesToRemove.add(instance.getId());
                } else {
                    instanceStore.save(instance);
                }
            }
        }

        for (String id : instancesToRemove) {
            instanceStore.findById(id).ifPresent(instanceService::markUnhealthyAndDeregister);
        }
    }
}
