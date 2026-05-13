package com.example.healthchecker.service;

import com.example.healthchecker.model.HealthStatus;
import com.example.healthchecker.model.ServiceRegistry;
import com.example.healthchecker.repository.HealthReportRepository;
import org.springframework.stereotype.Service;

import java.util.Set;

@Service
public class HealthStatusCalculator {

    private final HealthReportRepository healthReportRepository;

    public HealthStatusCalculator(HealthReportRepository healthReportRepository) {
        this.healthReportRepository = healthReportRepository;
    }

    public HealthStatus calculate(ServiceRegistry service) {
        if (service.isInMaintenance()) {
            return HealthStatus.MAINTENANCE;
        }

        Set<String> dependencies = service.getDependencies();
        
        if (dependencies.isEmpty()) {
            HealthStatus selfReported = healthReportRepository.getReportedStatusOrDefault(service.getName());
            if (selfReported == HealthStatus.MAINTENANCE) {
                return HealthStatus.MAINTENANCE;
            }
            return selfReported;
        }

        int totalDependencies = dependencies.size();
        int downCount = 0;
        boolean hasHealthy = false;
        boolean hasProblem = false;

        for (String depName : dependencies) {
            HealthStatus depStatus = healthReportRepository.getReportedStatusOrDefault(depName);
            
            if (depStatus == HealthStatus.DOWN) {
                downCount++;
                hasProblem = true;
            } else if (depStatus == HealthStatus.DEGRADED) {
                hasProblem = true;
            } else if (depStatus == HealthStatus.HEALTHY) {
                hasHealthy = true;
            }
        }

        if (downCount == totalDependencies) {
            return HealthStatus.DOWN;
        }

        if (hasProblem) {
            return HealthStatus.DEGRADED;
        }

        return HealthStatus.HEALTHY;
    }
}
