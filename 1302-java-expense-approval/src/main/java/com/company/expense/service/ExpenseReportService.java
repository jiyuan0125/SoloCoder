package com.company.expense.service;

import com.company.expense.dto.ApprovalRecordResponse;
import com.company.expense.dto.ApprovalRequest;
import com.company.expense.dto.ExpenseReportRequest;
import com.company.expense.dto.ExpenseReportResponse;
import com.company.expense.entity.ApprovalRecord;
import com.company.expense.entity.Employee;
import com.company.expense.entity.ExpenseReport;
import com.company.expense.enums.ApprovalAction;
import com.company.expense.enums.ApprovalLevel;
import com.company.expense.enums.ExpenseStatus;
import com.company.expense.exception.BusinessException;
import com.company.expense.exception.ResourceNotFoundException;
import com.company.expense.repository.ApprovalRecordRepository;
import com.company.expense.repository.ExpenseReportRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.LocalTime;
import java.util.List;
import java.util.UUID;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class ExpenseReportService {

    private final ExpenseReportRepository expenseReportRepository;
    private final ApprovalRecordRepository approvalRecordRepository;
    private final EmployeeService employeeService;
    private final ApprovalService approvalService;

    @Transactional
    public ExpenseReportResponse submitExpenseReport(ExpenseReportRequest request) {
        log.info("提交报销单，员工: {}", request.getApplicantEmployeeId());

        validateAmount(request.getAmount());

        Employee applicant = employeeService.getEmployeeEntity(request.getApplicantEmployeeId());

        ExpenseReport report;
        if (request.getId() != null) {
            report = expenseReportRepository.findById(request.getId())
                    .orElseThrow(() -> new ResourceNotFoundException("报销单不存在，ID: " + request.getId()));

            if (report.getStatus() != ExpenseStatus.REJECTED) {
                throw new BusinessException("只有已驳回的报销单才能重新提交");
            }

            if (!report.getApplicant().getEmployeeId().equals(request.getApplicantEmployeeId())) {
                throw new BusinessException("只能修改本人提交的报销单");
            }

            report.setExpenseType(request.getExpenseType());
            report.setAmount(request.getAmount());
            report.setDescription(request.getDescription());
            report.setRejectReason(null);
        } else {
            report = new ExpenseReport();
            report.setReportNo(generateReportNo());
            report.setApplicant(applicant);
            report.setExpenseType(request.getExpenseType());
            report.setAmount(request.getAmount());
            report.setDescription(request.getDescription());
        }

        ApprovalLevel initialLevel = ApprovalLevel.SUPERVISOR;
        Employee initialApprover = approvalService.determineApprover(applicant, initialLevel);

        report.setStatus(ExpenseStatus.PENDING);
        report.setCurrentApprovalLevel(initialLevel);
        report.setCurrentApprover(initialApprover);
        report.setSubmittedAt(LocalDateTime.now());
        report.setApprovedAt(null);

        report = expenseReportRepository.save(report);

        log.info("报销单 {} 提交成功，当前审批人: {}", 
                report.getReportNo(), initialApprover.getName());

        return ExpenseReportResponse.fromEntity(report);
    }

    @Transactional
    public ExpenseReportResponse processApproval(Long reportId, ApprovalRequest request) {
        log.info("处理审批，报销单 ID: {}, 审批人: {}", reportId, request.getApproverEmployeeId());

        ExpenseReport report = expenseReportRepository.findById(reportId)
                .orElseThrow(() -> new ResourceNotFoundException("报销单不存在，ID: " + reportId));

        if (report.getStatus() != ExpenseStatus.PENDING && report.getStatus() != ExpenseStatus.APPROVING) {
            throw new BusinessException("该报销单当前状态不允许审批操作");
        }

        Employee approver = employeeService.getEmployeeEntity(request.getApproverEmployeeId());

        if (report.getCurrentApprover() == null || 
            !report.getCurrentApprover().getEmployeeId().equals(approver.getEmployeeId())) {
            throw new BusinessException("您不是当前审批人，无权进行此操作");
        }

        if (approver.getEmployeeId().equals(report.getApplicant().getEmployeeId())) {
            throw new BusinessException("审批人不能是申请人本人");
        }

        ApprovalAction action;
        try {
            action = ApprovalAction.valueOf(request.getAction().toUpperCase());
        } catch (IllegalArgumentException e) {
            throw new BusinessException("无效的审批动作，只能是 APPROVE 或 REJECT");
        }

        ApprovalRecord record = new ApprovalRecord();
        record.setExpenseReport(report);
        record.setApprover(approver);
        record.setApprovalLevel(report.getCurrentApprovalLevel());
        record.setAction(action);
        record.setComment(request.getComment());
        approvalRecordRepository.save(record);

        if (action == ApprovalAction.REJECT) {
            report.setStatus(ExpenseStatus.REJECTED);
            report.setRejectReason(request.getComment());
            report.setCurrentApprover(null);
            expenseReportRepository.save(report);
            log.info("报销单 {} 已被驳回", report.getReportNo());
            return ExpenseReportResponse.fromEntity(report);
        }

        return processApprovalPassed(report);
    }

    private ExpenseReportResponse processApprovalPassed(ExpenseReport report) {
        ApprovalLevel currentLevel = report.getCurrentApprovalLevel();
        ApprovalLevel targetLevel = ApprovalLevel.getByAmount(report.getAmount().doubleValue());

        if (currentLevel == targetLevel || !approvalService.hasMoreApprovalLevels(currentLevel)) {
            report.setStatus(ExpenseStatus.APPROVED);
            report.setCurrentApprover(null);
            report.setApprovedAt(LocalDateTime.now());
            expenseReportRepository.save(report);
            log.info("报销单 {} 已全部通过审批", report.getReportNo());
            return ExpenseReportResponse.fromEntity(report);
        }

        ApprovalLevel nextLevel = approvalService.getNextApprovalLevel(currentLevel);
        if (nextLevel == null || nextLevel.compareTo(targetLevel) > 0) {
            report.setStatus(ExpenseStatus.APPROVED);
            report.setCurrentApprover(null);
            report.setApprovedAt(LocalDateTime.now());
            expenseReportRepository.save(report);
            log.info("报销单 {} 已全部通过审批", report.getReportNo());
            return ExpenseReportResponse.fromEntity(report);
        }

        Employee nextApprover = approvalService.determineApprover(report.getApplicant(), nextLevel);
        report.setStatus(ExpenseStatus.APPROVING);
        report.setCurrentApprovalLevel(nextLevel);
        report.setCurrentApprover(nextApprover);
        expenseReportRepository.save(report);

        log.info("报销单 {} 流转到下一级审批，审批人: {}", 
                report.getReportNo(), nextApprover.getName());

        return ExpenseReportResponse.fromEntity(report);
    }

    public ExpenseReportResponse getExpenseReport(Long id) {
        log.debug("查询报销单 ID: {}", id);

        ExpenseReport report = expenseReportRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("报销单不存在，ID: " + id));

        return ExpenseReportResponse.fromEntity(report);
    }

    public ExpenseReportResponse getExpenseReportByNo(String reportNo) {
        log.debug("查询报销单号: {}", reportNo);

        ExpenseReport report = expenseReportRepository.findByReportNo(reportNo)
                .orElseThrow(() -> new ResourceNotFoundException("报销单不存在: " + reportNo));

        return ExpenseReportResponse.fromEntity(report);
    }

    public Page<ExpenseReportResponse> getMyExpenseReports(
            String employeeId,
            ExpenseStatus status,
            String expenseType,
            LocalDate startDate,
            LocalDate endDate,
            int page,
            int size) {

        log.debug("查询员工 {} 的报销单", employeeId);

        Employee employee = employeeService.getEmployeeEntity(employeeId);
        Pageable pageable = PageRequest.of(page, size, Sort.by(Sort.Direction.DESC, "submittedAt"));

        com.company.expense.enums.ExpenseType type = null;
        if (expenseType != null && !expenseType.isEmpty()) {
            try {
                type = com.company.expense.enums.ExpenseType.valueOf(expenseType.toUpperCase());
            } catch (IllegalArgumentException e) {
                throw new BusinessException("无效的报销类型: " + expenseType);
            }
        }

        LocalDateTime start = startDate != null ? startDate.atStartOfDay() : null;
        LocalDateTime end = endDate != null ? endDate.atTime(LocalTime.MAX) : null;

        Page<ExpenseReport> reports = expenseReportRepository.findByApplicantWithFilters(
                employee, status, type, start, end, pageable);

        return reports.map(ExpenseReportResponse::fromEntity);
    }

    public Page<ExpenseReportResponse> getAllExpenseReports(
            ExpenseStatus status,
            String expenseType,
            LocalDate startDate,
            LocalDate endDate,
            int page,
            int size) {

        log.debug("查询所有报销单");

        Pageable pageable = PageRequest.of(page, size, Sort.by(Sort.Direction.DESC, "submittedAt"));

        com.company.expense.enums.ExpenseType type = null;
        if (expenseType != null && !expenseType.isEmpty()) {
            try {
                type = com.company.expense.enums.ExpenseType.valueOf(expenseType.toUpperCase());
            } catch (IllegalArgumentException e) {
                throw new BusinessException("无效的报销类型: " + expenseType);
            }
        }

        LocalDateTime start = startDate != null ? startDate.atStartOfDay() : null;
        LocalDateTime end = endDate != null ? endDate.atTime(LocalTime.MAX) : null;

        Page<ExpenseReport> reports = expenseReportRepository.findAllWithFilters(
                status, type, start, end, pageable);

        return reports.map(ExpenseReportResponse::fromEntity);
    }

    public Page<ExpenseReportResponse> getPendingExpenseReports(
            String approverEmployeeId,
            int page,
            int size) {

        log.debug("查询审批人 {} 的待审批报销单", approverEmployeeId);

        Employee approver = employeeService.getEmployeeEntity(approverEmployeeId);
        Pageable pageable = PageRequest.of(page, size, Sort.by(Sort.Direction.DESC, "submittedAt"));

        Page<ExpenseReport> reports = expenseReportRepository.findPendingByApprover(approver, pageable);
        return reports.map(ExpenseReportResponse::fromEntity);
    }

    public List<ApprovalRecordResponse> getApprovalHistory(Long reportId) {
        log.debug("查询报销单 {} 的审批历史", reportId);

        ExpenseReport report = expenseReportRepository.findById(reportId)
                .orElseThrow(() -> new ResourceNotFoundException("报销单不存在，ID: " + reportId));

        return approvalRecordRepository.findByExpenseReportOrderByCreatedAt(report).stream()
                .map(ApprovalRecordResponse::fromEntity)
                .collect(Collectors.toList());
    }

    private void validateAmount(BigDecimal amount) {
        if (amount == null) {
            throw new BusinessException("报销金额不能为空");
        }
        if (amount.compareTo(BigDecimal.ZERO) <= 0) {
            throw new BusinessException("报销金额必须大于0");
        }
        if (amount.compareTo(new BigDecimal("99999.99")) > 0) {
            throw new BusinessException("报销金额不能超过99999.99元");
        }
    }

    private String generateReportNo() {
        return "EXP-" + System.currentTimeMillis() + "-" + UUID.randomUUID().toString().substring(0, 6).toUpperCase();
    }
}
