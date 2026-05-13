package com.example.healthchecker.service;

import com.example.healthchecker.model.HealthReport;
import com.example.healthchecker.model.HealthStatus;
import com.example.healthchecker.model.ServiceRegistry;
import com.example.healthchecker.repository.HealthReportRepository;
import com.example.healthchecker.repository.ServiceRegistryRepository;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.Collection;
import java.util.Set;

@Service
public class HealthCheckService {

    private final ServiceRegistryRepository serviceRegistryRepository;
    private final HealthReportRepository healthReportRepository;
    private final HealthStatusCalculator healthStatusCalculator;
    private final EventLogService eventLogService;

    public HealthCheckService(ServiceRegistryRepository serviceRegistryRepository,
                              HealthReportRepository healthReportRepository,
                              HealthStatusCalculator healthStatusCalculator,
                              EventLogService eventLogService) {
        this.serviceRegistryRepository = serviceRegistryRepository;
        this.healthReportRepository = healthReportRepository;
        this.healthStatusCalculator = healthStatusCalculator;
        this.eventLogService = eventLogService;
    }

    public ServiceRegistry getOrCreateService(String name) {
        return serviceRegistryRepository.getOrCreate(name);
    }

    public void addDependency(String serviceName, String dependencyServiceName) {
        ServiceRegistry service = serviceRegistryRepository.getOrCreate(serviceName);
        serviceRegistryRepository.getOrCreate(dependencyServiceName);
        service.getDependencies().add(dependencyServiceName);
        recalculateServiceStatus(service, "dependency added: " + dependencyServiceName);
    }

    public void removeDependency(String serviceName, String dependencyServiceName) {
        serviceRegistryRepository.findByName(serviceName).ifPresent(service -> {
            service.getDependencies().remove(dependencyServiceName);
            recalculateServiceStatus(service, "dependency removed: " + dependencyServiceName);
        });
    }

    public void reportStatus(String serviceName, HealthStatus status) {
        HealthReport report = new HealthReport();
        report.setServiceName(serviceName);
        report.setStatus(status);
        healthReportRepository.save(report);

        ServiceRegistry service = serviceRegistryRepository.getOrCreate(serviceName);
        recalculateServiceStatus(service, "status reported: " + status);

        Set<String> dependents = serviceRegistryRepository.findDependentServices(serviceName);
        for (String dependentName : dependents) {
            serviceRegistryRepository.findByName(dependentName).ifPresent(dep -> {
                recalculateServiceStatus(dep, "dependency changed: " + serviceName + " -> " + status);
            });
        }
    }

    public void setMaintenanceMode(String serviceName, boolean enabled) {
        ServiceRegistry service = serviceRegistryRepository.getOrCreate(serviceName);
        HealthStatus previousStatus = service.getStatus();
        boolean previousMaintenance = service.isInMaintenance();

        service.setInMaintenance(enabled);
        recalculateServiceStatus(service, "maintenance mode " + (enabled ? "enabled" : "disabled"));
    }

    public void recalculateAll(String triggerReason) {
        Collection<ServiceRegistry> allServices = serviceRegistryRepository.findAll();
        for (ServiceRegistry service : allServices) {
            recalculateServiceStatus(service, triggerReason);
        }
    }

    private void recalculateServiceStatus(ServiceRegistry service, String reason) {
        HealthStatus previousStatus = service.getStatus();
        HealthStatus newStatus = healthStatusCalculator.calculate(service);

        if (previousStatus != newStatus) {
            service.setStatus(newStatus);
            service.setLastStatusChangeTime(LocalDateTime.now());
            eventLogService.logStatusChange(service.getName(), previousStatus, newStatus, reason);
        }
    }

    public Collection<ServiceRegistry> getAllServices() {
        return serviceRegistryRepository.findAll();
    }
}
