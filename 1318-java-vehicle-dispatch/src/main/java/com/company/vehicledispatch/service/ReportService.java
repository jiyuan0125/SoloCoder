package com.company.vehicledispatch.service;

import com.company.vehicledispatch.entity.DispatchRecord;
import com.company.vehicledispatch.entity.Maintenance;
import com.company.vehicledispatch.entity.MonthlyReport;
import com.company.vehicledispatch.entity.Vehicle;
import com.company.vehicledispatch.exception.BusinessException;
import com.company.vehicledispatch.repository.DispatchRecordRepository;
import com.company.vehicledispatch.repository.MaintenanceRepository;
import com.company.vehicledispatch.repository.MonthlyReportRepository;
import com.company.vehicledispatch.repository.VehicleRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.YearMonth;
import java.util.*;
import java.util.stream.Collectors;

@Service
public class ReportService {

    @Autowired
    private MonthlyReportRepository monthlyReportRepository;

    @Autowired
    private VehicleRepository vehicleRepository;

    @Autowired
    private DispatchRecordRepository dispatchRecordRepository;

    @Autowired
    private MaintenanceRepository maintenanceRepository;

    private static final double USAGE_RATE_THRESHOLD = 0.30;

    public List<MonthlyReport> getMonthlyReports(Integer year, Integer month) {
        return monthlyReportRepository.findByReportYearAndReportMonth(year, month);
    }

    public Optional<MonthlyReport> getVehicleMonthlyReport(Long vehicleId, Integer year, Integer month) {
        return monthlyReportRepository.findByVehicleIdAndReportYearAndReportMonth(vehicleId, year, month);
    }

    public List<MonthlyReport> getReportsByVehicle(Long vehicleId) {
        return monthlyReportRepository.findByVehicleId(vehicleId);
    }

    @Transactional
    public List<MonthlyReport> generateMonthlyReports(Integer year, Integer month) {
        List<Vehicle> vehicles = vehicleRepository.findAll();
        List<MonthlyReport> reports = new ArrayList<>();

        for (Vehicle vehicle : vehicles) {
            Optional<MonthlyReport> existingReport = monthlyReportRepository
                    .findByVehicleIdAndReportYearAndReportMonth(vehicle.getId(), year, month);

            if (existingReport.isPresent()) {
                reports.add(existingReport.get());
                continue;
            }

            MonthlyReport report = generateVehicleMonthlyReport(vehicle, year, month);
            reports.add(monthlyReportRepository.save(report));
        }

        return reports;
    }

    @Transactional
    public MonthlyReport generateVehicleMonthlyReport(Long vehicleId, Integer year, Integer month) {
        Vehicle vehicle = vehicleRepository.findById(vehicleId)
                .orElseThrow(() -> new BusinessException("车辆不存在: " + vehicleId));

        Optional<MonthlyReport> existingReport = monthlyReportRepository
                .findByVehicleIdAndReportYearAndReportMonth(vehicleId, year, month);

        if (existingReport.isPresent()) {
            return existingReport.get();
        }

        MonthlyReport report = generateVehicleMonthlyReport(vehicle, year, month);
        return monthlyReportRepository.save(report);
    }

    private MonthlyReport generateVehicleMonthlyReport(Vehicle vehicle, Integer year, Integer month) {
        YearMonth yearMonth = YearMonth.of(year, month);
        LocalDate startDate = yearMonth.atDay(1);
        LocalDate endDate = yearMonth.atEndOfMonth();
        int daysInMonth = yearMonth.lengthOfMonth();

        LocalDateTime startDateTime = startDate.atStartOfDay();
        LocalDateTime endDateTime = endDate.plusDays(1).atStartOfDay();

        List<DispatchRecord> records = dispatchRecordRepository.findByVehicleIdAndMonth(
                vehicle.getId(), startDateTime, endDateTime);

        Set<LocalDate> usageDates = records.stream()
                .map(r -> r.getActualStartDateTime().toLocalDate())
                .collect(Collectors.toSet());
        int usageDays = usageDates.size();

        double totalMileage = records.stream()
                .filter(r -> r.getActualDistance() != null)
                .mapToDouble(DispatchRecord::getActualDistance)
                .sum();

        double totalFuelConsumption = records.stream()
                .filter(r -> r.getActualFuelConsumption() != null)
                .mapToDouble(DispatchRecord::getActualFuelConsumption)
                .sum();

        List<Maintenance> maintenances = maintenanceRepository.findByVehicleIdAndMonth(
                vehicle.getId(), startDate, endDate.plusDays(1));

        double maintenanceCost = maintenances.stream()
                .filter(m -> m.getCost() != null)
                .mapToDouble(Maintenance::getCost)
                .sum();

        double usageRate = (double) usageDays / daysInMonth;

        String recommendation = generateRecommendation(usageRate, totalMileage);

        MonthlyReport report = new MonthlyReport();
        report.setVehicle(vehicle);
        report.setReportYear(year);
        report.setReportMonth(month);
        report.setUsageDays(usageDays);
        report.setTotalMileage(totalMileage);
        report.setTotalFuelConsumption(totalFuelConsumption);
        report.setMaintenanceCost(maintenanceCost);
        report.setUsageRate(usageRate);
        report.setRecommendation(recommendation);
        report.setCreatedAt(LocalDateTime.now());

        return report;
    }

    private String generateRecommendation(double usageRate, double totalMileage) {
        List<String> recommendations = new ArrayList<>();

        if (usageRate < USAGE_RATE_THRESHOLD) {
            recommendations.add("使用率较低，建议调配备用或考虑处置");
        }

        if (totalMileage > 5000) {
            recommendations.add("本月行驶里程较高，建议检查保养");
        }

        if (recommendations.isEmpty()) {
            recommendations.add("使用状况正常");
        }

        return String.join("; ", recommendations);
    }
}
