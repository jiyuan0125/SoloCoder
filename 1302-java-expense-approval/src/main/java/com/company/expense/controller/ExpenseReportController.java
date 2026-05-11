package com.company.expense.controller;

import com.company.expense.dto.ApprovalRecordResponse;
import com.company.expense.dto.ApprovalRequest;
import com.company.expense.dto.ExpenseReportRequest;
import com.company.expense.dto.ExpenseReportResponse;
import com.company.expense.enums.ExpenseStatus;
import com.company.expense.service.ExpenseReportService;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.time.LocalDate;
import java.util.List;

@RestController
@RequestMapping("/api/expense-reports")
@RequiredArgsConstructor
public class ExpenseReportController {

    private final ExpenseReportService expenseReportService;

    @PostMapping("/submit")
    public ResponseEntity<ExpenseReportResponse> submitExpenseReport(
            @Valid @RequestBody ExpenseReportRequest request) {
        ExpenseReportResponse response = expenseReportService.submitExpenseReport(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @GetMapping("/{id}")
    public ResponseEntity<ExpenseReportResponse> getExpenseReport(@PathVariable Long id) {
        ExpenseReportResponse response = expenseReportService.getExpenseReport(id);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/by-no/{reportNo}")
    public ResponseEntity<ExpenseReportResponse> getExpenseReportByNo(@PathVariable String reportNo) {
        ExpenseReportResponse response = expenseReportService.getExpenseReportByNo(reportNo);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/my")
    public ResponseEntity<Page<ExpenseReportResponse>> getMyExpenseReports(
            @RequestParam String employeeId,
            @RequestParam(required = false) ExpenseStatus status,
            @RequestParam(required = false) String expenseType,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate startDate,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate endDate,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "10") int size) {
        
        Page<ExpenseReportResponse> response = expenseReportService.getMyExpenseReports(
                employeeId, status, expenseType, startDate, endDate, page, size);
        return ResponseEntity.ok(response);
    }

    @GetMapping
    public ResponseEntity<Page<ExpenseReportResponse>> getAllExpenseReports(
            @RequestParam(required = false) ExpenseStatus status,
            @RequestParam(required = false) String expenseType,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate startDate,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate endDate,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "10") int size) {
        
        Page<ExpenseReportResponse> response = expenseReportService.getAllExpenseReports(
                status, expenseType, startDate, endDate, page, size);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/pending")
    public ResponseEntity<Page<ExpenseReportResponse>> getPendingExpenseReports(
            @RequestParam String approverEmployeeId,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "10") int size) {
        
        Page<ExpenseReportResponse> response = expenseReportService.getPendingExpenseReports(
                approverEmployeeId, page, size);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/{id}/approve")
    public ResponseEntity<ExpenseReportResponse> processApproval(
            @PathVariable Long id,
            @Valid @RequestBody ApprovalRequest request) {
        ExpenseReportResponse response = expenseReportService.processApproval(id, request);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/{id}/history")
    public ResponseEntity<List<ApprovalRecordResponse>> getApprovalHistory(@PathVariable Long id) {
        List<ApprovalRecordResponse> response = expenseReportService.getApprovalHistory(id);
        return ResponseEntity.ok(response);
    }
}
