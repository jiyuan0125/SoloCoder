package com.company.leave.controller;

import com.company.leave.dto.ApiResponse;
import com.company.leave.dto.ApproveLeaveRequest;
import com.company.leave.dto.CreateLeaveRequest;
import com.company.leave.dto.LeaveRecordDTO;
import com.company.leave.entity.Employee;
import com.company.leave.entity.LeaveRecord;
import com.company.leave.enums.ErrorCode;
import com.company.leave.service.EmployeeService;
import com.company.leave.service.LeaveService;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/api/leaves")
public class LeaveController {
    private final LeaveService leaveService;
    private final EmployeeService employeeService;

    public LeaveController(LeaveService leaveService, EmployeeService employeeService) {
        this.leaveService = leaveService;
        this.employeeService = employeeService;
    }

    @PostMapping
    public ApiResponse<LeaveRecordDTO> createLeave(@RequestBody CreateLeaveRequest request) {
        try {
            LeaveRecord record = leaveService.createLeave(request);
            return ApiResponse.success(toDTO(record));
        } catch (IllegalArgumentException e) {
            return ApiResponse.error(ErrorCode.INVALID_REQUEST, e.getMessage());
        }
    }

    @PostMapping("/approve")
    public ApiResponse<LeaveRecordDTO> approveLeave(@RequestBody ApproveLeaveRequest request) {
        try {
            LeaveRecord record = leaveService.approveLeave(
                    request.getLeaveId(),
                    request.getManagerId(),
                    request.isApproved(),
                    request.getComment()
            );
            return ApiResponse.success(toDTO(record));
        } catch (IllegalArgumentException e) {
            return ApiResponse.error(ErrorCode.INVALID_REQUEST, e.getMessage());
        }
    }

    @GetMapping("/{id}")
    public ApiResponse<LeaveRecordDTO> getLeaveRecord(@PathVariable Long id) {
        return leaveService.getLeaveRecord(id)
                .map(record -> ApiResponse.success(toDTO(record)))
                .orElse(ApiResponse.error(ErrorCode.LEAVE_NOT_FOUND));
    }

    @GetMapping("/employee/{employeeId}")
    public ApiResponse<List<LeaveRecordDTO>> getEmployeeLeaves(@PathVariable Long employeeId) {
        List<LeaveRecordDTO> dtos = leaveService.getLeaveRecordsByEmployee(employeeId).stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
        return ApiResponse.success(dtos);
    }

    @GetMapping
    public ApiResponse<List<LeaveRecordDTO>> getAllLeaves() {
        List<LeaveRecordDTO> dtos = leaveService.getAllLeaveRecords().stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
        return ApiResponse.success(dtos);
    }

    private LeaveRecordDTO toDTO(LeaveRecord record) {
        LeaveRecordDTO dto = new LeaveRecordDTO();
        dto.setId(record.getId());
        dto.setEmployeeId(record.getEmployeeId());
        
        Optional<Employee> employeeOpt = employeeService.getEmployee(record.getEmployeeId());
        employeeOpt.ifPresent(e -> dto.setEmployeeName(e.getName()));
        
        dto.setLeaveType(record.getLeaveType());
        dto.setStartDate(record.getStartDate());
        dto.setEndDate(record.getEndDate());
        dto.setActualDays(record.getActualDays());
        dto.setReason(record.getReason());
        dto.setAttachmentName(record.getAttachmentName());
        dto.setStatus(record.getStatus());
        dto.setApprovedBy(record.getApprovedBy());
        dto.setApprovalComment(record.getApprovalComment());
        dto.setCreatedAt(record.getCreatedAt());
        dto.setUpdatedAt(record.getUpdatedAt());
        return dto;
    }
}
