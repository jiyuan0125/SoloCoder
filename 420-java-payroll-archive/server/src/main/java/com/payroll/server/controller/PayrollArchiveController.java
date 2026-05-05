package com.payroll.server.controller;

import com.payroll.common.dto.*;
import com.payroll.common.enums.ErrorCode;
import com.payroll.common.response.ApiResponse;
import com.payroll.server.service.PayrollArchiveService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/api/archives")
public class PayrollArchiveController {

    private final PayrollArchiveService service;

    @Autowired
    public PayrollArchiveController(PayrollArchiveService service) {
        this.service = service;
    }

    @PostMapping
    public ApiResponse<PayrollArchiveDTO> createArchive(@RequestBody ArchiveRequestDTO request) {
        try {
            validateArchiveRequest(request);
            PayrollArchiveDTO result = service.createArchive(request);
            return ApiResponse.success(result);
        } catch (IllegalArgumentException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM.getCode(), e.getMessage());
        } catch (Exception e) {
            return ApiResponse.error(ErrorCode.INTERNAL_ERROR);
        }
    }

    @PostMapping("/{archiveId}/confirm")
    public ApiResponse<PayrollArchiveDTO> confirmArchive(@PathVariable String archiveId) {
        try {
            PayrollArchiveDTO result = service.confirmArchive(archiveId);
            return ApiResponse.success(result);
        } catch (IllegalArgumentException e) {
            if (e.getMessage().contains("不存在")) {
                return ApiResponse.error(ErrorCode.ARCHIVE_NOT_FOUND);
            }
            return ApiResponse.error(ErrorCode.INVALID_PARAM.getCode(), e.getMessage());
        } catch (Exception e) {
            return ApiResponse.error(ErrorCode.INTERNAL_ERROR);
        }
    }

    @PostMapping("/{archiveId}/remarks")
    public ApiResponse<PayrollArchiveDTO> addRemark(
            @PathVariable String archiveId,
            @RequestBody RemarkRequestDTO request) {
        try {
            PayrollArchiveDTO result = service.addRemark(archiveId, request.getRemark());
            return ApiResponse.success(result);
        } catch (IllegalArgumentException e) {
            if (e.getMessage().contains("不存在")) {
                return ApiResponse.error(ErrorCode.ARCHIVE_NOT_FOUND);
            }
            return ApiResponse.error(ErrorCode.INVALID_PARAM.getCode(), e.getMessage());
        } catch (Exception e) {
            return ApiResponse.error(ErrorCode.INTERNAL_ERROR);
        }
    }

    @GetMapping("/{archiveId}")
    public ApiResponse<PayrollArchiveDTO> getByArchiveId(@PathVariable String archiveId) {
        try {
            PayrollArchiveDTO result = service.getByArchiveId(archiveId);
            return ApiResponse.success(result);
        } catch (IllegalArgumentException e) {
            return ApiResponse.error(ErrorCode.ARCHIVE_NOT_FOUND);
        } catch (Exception e) {
            return ApiResponse.error(ErrorCode.INTERNAL_ERROR);
        }
    }

    @GetMapping
    public ApiResponse<List<PayrollArchiveDTO>> query(
            @RequestParam(required = false) String employeeId,
            @RequestParam(required = false) Integer year,
            @RequestParam(required = false) Integer month) {
        try {
            List<PayrollArchiveDTO> result;
            
            if (employeeId != null && year != null && month != null) {
                result = service.queryByEmployeeAndYear(employeeId, year).stream()
                        .filter(a -> a.getMonth() == month)
                        .collect(Collectors.toList());
            } else if (employeeId != null && year != null) {
                result = service.queryByEmployeeAndYear(employeeId, year);
            } else if (employeeId != null) {
                result = service.queryByEmployee(employeeId);
            } else if (year != null && month != null) {
                result = service.queryByYearMonth(year, month);
            } else if (year != null) {
                result = service.queryByYear(year);
            } else {
                return ApiResponse.error(ErrorCode.INVALID_PARAM.getCode(), "请提供至少一个查询条件");
            }
            
            return ApiResponse.success(result);
        } catch (Exception e) {
            return ApiResponse.error(ErrorCode.INTERNAL_ERROR);
        }
    }

    @GetMapping("/years/{year}/report")
    public ApiResponse<AnnualReportDTO> getAnnualReport(@PathVariable int year) {
        try {
            AnnualReportDTO result = service.generateAnnualReport(year);
            return ApiResponse.success(result);
        } catch (Exception e) {
            return ApiResponse.error(ErrorCode.INTERNAL_ERROR);
        }
    }

    @GetMapping("/years/{year}/report/export")
    public ApiResponse<AnnualReportExportDTO> exportAnnualReport(@PathVariable int year) {
        try {
            AnnualReportExportDTO result = service.exportAnnualReport(year);
            return ApiResponse.success(result);
        } catch (Exception e) {
            return ApiResponse.error(ErrorCode.INTERNAL_ERROR);
        }
    }

    @GetMapping("/compare")
    public ApiResponse<YearComparisonDTO> compareYears(
            @RequestParam int baseYear,
            @RequestParam int compareYear) {
        try {
            YearComparisonDTO result = service.compareYears(baseYear, compareYear);
            return ApiResponse.success(result);
        } catch (Exception e) {
            return ApiResponse.error(ErrorCode.INTERNAL_ERROR);
        }
    }

    private void validateArchiveRequest(ArchiveRequestDTO request) {
        if (request.getEmployeeId() == null || request.getEmployeeId().trim().isEmpty()) {
            throw new IllegalArgumentException("员工ID不能为空");
        }
        if (request.getEmployeeName() == null || request.getEmployeeName().trim().isEmpty()) {
            throw new IllegalArgumentException("员工姓名不能为空");
        }
        if (request.getYear() < 2000 || request.getYear() > 2100) {
            throw new IllegalArgumentException("年份无效");
        }
        if (request.getMonth() < 1 || request.getMonth() > 12) {
            throw new IllegalArgumentException("月份无效");
        }
        if (request.getBaseSalary() == null || request.getBaseSalary().compareTo(java.math.BigDecimal.ZERO) < 0) {
            throw new IllegalArgumentException("基本工资无效");
        }
    }
}
