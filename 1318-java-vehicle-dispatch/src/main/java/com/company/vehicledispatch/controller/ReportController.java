package com.company.vehicledispatch.controller;

import com.company.vehicledispatch.entity.MonthlyReport;
import com.company.vehicledispatch.service.ReportService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/reports")
public class ReportController {

    @Autowired
    private ReportService reportService;

    @GetMapping("/monthly")
    public ResponseEntity<List<MonthlyReport>> getMonthlyReports(
            @RequestParam Integer year,
            @RequestParam Integer month) {
        return ResponseEntity.ok(reportService.getMonthlyReports(year, month));
    }

    @GetMapping("/vehicle/{vehicleId}/monthly")
    public ResponseEntity<MonthlyReport> getVehicleMonthlyReport(
            @PathVariable Long vehicleId,
            @RequestParam Integer year,
            @RequestParam Integer month) {
        return reportService.getVehicleMonthlyReport(vehicleId, year, month)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/vehicle/{vehicleId}")
    public ResponseEntity<List<MonthlyReport>> getReportsByVehicle(@PathVariable Long vehicleId) {
        return ResponseEntity.ok(reportService.getReportsByVehicle(vehicleId));
    }

    @PostMapping("/monthly/generate")
    public ResponseEntity<List<MonthlyReport>> generateMonthlyReports(
            @RequestParam Integer year,
            @RequestParam Integer month) {
        List<MonthlyReport> reports = reportService.generateMonthlyReports(year, month);
        return ResponseEntity.ok(reports);
    }

    @PostMapping("/vehicle/{vehicleId}/monthly/generate")
    public ResponseEntity<MonthlyReport> generateVehicleMonthlyReport(
            @PathVariable Long vehicleId,
            @RequestParam Integer year,
            @RequestParam Integer month) {
        MonthlyReport report = reportService.generateVehicleMonthlyReport(vehicleId, year, month);
        return ResponseEntity.ok(report);
    }
}
