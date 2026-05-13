package com.example.healthchecker.repository;

import com.example.healthchecker.model.HealthReport;
import com.example.healthchecker.model.HealthStatus;
import org.springframework.stereotype.Repository;

import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;

@Repository
public class HealthReportRepository {

    private final Map<String, HealthReport> reports = new ConcurrentHashMap<>();

    public void save(HealthReport report) {
        reports.put(report.getServiceName(), report);
    }

    public Optional<HealthReport> findLatestByServiceName(String serviceName) {
        return Optional.ofNullable(reports.get(serviceName));
    }

    public HealthStatus getReportedStatusOrDefault(String serviceName) {
        return reports.getOrDefault(serviceName, createDefaultReport(serviceName)).getStatus();
    }

    private HealthReport createDefaultReport(String serviceName) {
        HealthReport report = new HealthReport();
        report.setServiceName(serviceName);
        report.setStatus(HealthStatus.HEALTHY);
        return report;
    }
}
