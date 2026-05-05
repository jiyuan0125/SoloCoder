package com.reimbursement.server.controller;

import com.reimbursement.common.dto.*;
import com.reimbursement.server.service.MonthlyReportService;
import com.reimbursement.server.service.ReimbursementService;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/reimbursement")
public class ReimbursementController {
    private final ReimbursementService reimbursementService;
    private final MonthlyReportService monthlyReportService;

    public ReimbursementController(ReimbursementService reimbursementService, 
                                     MonthlyReportService monthlyReportService) {
        this.reimbursementService = reimbursementService;
        this.monthlyReportService = monthlyReportService;
    }

    @PostMapping("/create")
    public ApiResponse<ReimbursementDetailDTO> createReimbursement(@RequestBody CreateReimbursementRequest request) {
        return reimbursementService.createReimbursement(request);
    }

    @PostMapping("/submit/{id}")
    public ApiResponse<ReimbursementDetailDTO> submitReimbursement(@PathVariable String id) {
        return reimbursementService.submitReimbursement(id);
    }

    @PutMapping("/update")
    public ApiResponse<ReimbursementDetailDTO> updateReimbursement(@RequestBody UpdateReimbursementRequest request) {
        return reimbursementService.updateReimbursement(request);
    }

    @PostMapping("/approve")
    public ApiResponse<ReimbursementDetailDTO> approve(@RequestBody ApprovalRequest request) {
        return reimbursementService.approve(request);
    }

    @PostMapping("/final-review")
    public ApiResponse<ReimbursementDetailDTO> finalReview(@RequestBody FinalReviewRequest request) {
        return reimbursementService.finalReview(request);
    }

    @GetMapping("/{id}")
    public ApiResponse<ReimbursementDetailDTO> getReimbursement(@PathVariable String id) {
        return reimbursementService.getReimbursement(id);
    }

    @GetMapping("/list")
    public ApiResponse<List<ReimbursementDetailDTO>> getAllReimbursements() {
        return reimbursementService.getAllReimbursements();
    }

    @GetMapping("/cost-centers")
    public ApiResponse<List<CostCenterDTO>> getAllCostCenters() {
        return reimbursementService.getAllCostCenters();
    }

    @GetMapping("/report/{year}/{month}/{costCenterId}")
    public ApiResponse<MonthlyReportDTO> getMonthlyReport(
            @PathVariable int year,
            @PathVariable int month,
            @PathVariable String costCenterId) {
        MonthlyReportDTO report = monthlyReportService.getMonthlyReport(year, month, costCenterId);
        if (report == null) {
            return ApiResponse.error(404, "报表不存在");
        }
        return ApiResponse.success(report);
    }

    @GetMapping("/report/{year}/{month}")
    public ApiResponse<List<MonthlyReportDTO>> getAllMonthlyReports(
            @PathVariable int year,
            @PathVariable int month) {
        return ApiResponse.success(monthlyReportService.getAllMonthlyReports(year, month));
    }

    @GetMapping("/budget-warning/{costCenterId}")
    public ApiResponse<Boolean> checkBudgetWarning(@PathVariable String costCenterId) {
        return ApiResponse.success(monthlyReportService.checkBudgetWarning(costCenterId));
    }
}
