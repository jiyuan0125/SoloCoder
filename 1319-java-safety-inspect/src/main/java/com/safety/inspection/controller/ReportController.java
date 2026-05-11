package com.safety.inspection.controller;

import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.safety.inspection.common.Result;
import com.safety.inspection.dto.DashboardStatsVO;
import com.safety.inspection.dto.MonthlyReportVO;
import com.safety.inspection.entity.MonthlyReport;
import com.safety.inspection.service.ReportService;
import lombok.RequiredArgsConstructor;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/reports")
@RequiredArgsConstructor
public class ReportController {

    private final ReportService reportService;

    @GetMapping("/dashboard")
    public Result<DashboardStatsVO> getDashboardStats() {
        return Result.success(reportService.getDashboardStats());
    }

    @GetMapping("/monthly/{year}/{month}")
    @PreAuthorize("hasRole('ADMIN') or hasRole('DEPT_LEADER') or hasRole('SAFETY_DIRECTOR') or hasRole('FACTORY_MANAGER')")
    public Result<MonthlyReportVO> getReportByMonth(@PathVariable int year, @PathVariable int month) {
        return Result.success(reportService.getReportByMonth(year, month));
    }

    @PostMapping("/monthly/{year}/{month}/generate")
    @PreAuthorize("hasRole('ADMIN')")
    public Result<MonthlyReportVO> generateReport(@PathVariable int year, @PathVariable int month) {
        return Result.success(reportService.generateReportForMonth(year, month));
    }

    @GetMapping("/monthly/page")
    @PreAuthorize("hasRole('ADMIN') or hasRole('DEPT_LEADER') or hasRole('SAFETY_DIRECTOR') or hasRole('FACTORY_MANAGER')")
    public Result<Page<MonthlyReport>> getReportPage(
            @RequestParam(defaultValue = "1") int pageNum,
            @RequestParam(defaultValue = "10") int pageSize) {
        return Result.success(reportService.getReportPage(pageNum, pageSize));
    }
}
