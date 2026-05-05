package com.company.leave.service;

import com.company.leave.dto.CreateLeaveRequest;
import com.company.leave.entity.Employee;
import com.company.leave.entity.LeaveRecord;
import com.company.leave.enums.LeaveStatus;
import com.company.leave.enums.LeaveType;
import com.company.leave.repository.LeaveRecordRepository;
import com.company.leave.util.DateCalculator;
import org.springframework.stereotype.Service;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;

@Service
public class LeaveService {
    private final LeaveRecordRepository leaveRecordRepository;
    private final EmployeeService employeeService;
    private final AnnualLeaveService annualLeaveService;

    public LeaveService(LeaveRecordRepository leaveRecordRepository,
                        EmployeeService employeeService,
                        AnnualLeaveService annualLeaveService) {
        this.leaveRecordRepository = leaveRecordRepository;
        this.employeeService = employeeService;
        this.annualLeaveService = annualLeaveService;
    }

    public LeaveRecord createLeave(CreateLeaveRequest request) {
        validateRequest(request);
        
        Optional<Employee> employeeOpt = employeeService.getEmployee(request.getEmployeeId());
        if (employeeOpt.isEmpty()) {
            throw new IllegalArgumentException("员工不存在");
        }
        
        Employee employee = employeeOpt.get();
        
        LeaveType leaveType = request.getLeaveType();
        LocalDate startDate = request.getStartDate();
        LocalDate endDate = request.getEndDate();
        
        if (leaveType.isRequiresAttachment()) {
            String attachment = request.getAttachmentName();
            if (attachment == null || attachment.trim().isEmpty()) {
                throw new IllegalArgumentException("该请假类型必须提供附件");
            }
        }
        
        List<LeaveRecord> overlapping = leaveRecordRepository.findOverlappingByEmployeeId(
                employee.getId(), startDate, endDate);
        if (!overlapping.isEmpty()) {
            throw new IllegalArgumentException("同一时间段已存在请假记录");
        }
        
        int actualDays = DateCalculator.calculateLeaveDays(startDate, endDate, leaveType);
        
        if (LeaveType.ANNUAL.equals(leaveType)) {
            int available = annualLeaveService.getTotalAvailableAnnualLeave(employee);
            if (available < actualDays) {
                throw new IllegalArgumentException("年假天数不足，可用：" + available + "天，申请：" + actualDays + "天");
            }
        }
        
        validateFixedLeaveQuota(leaveType, actualDays);
        
        LeaveRecord record = new LeaveRecord();
        record.setEmployeeId(employee.getId());
        record.setLeaveType(leaveType);
        record.setStartDate(startDate);
        record.setEndDate(endDate);
        record.setActualDays(actualDays);
        record.setReason(request.getReason());
        record.setAttachmentName(request.getAttachmentName());
        record.setStatus(LeaveStatus.PENDING);
        
        return leaveRecordRepository.save(record);
    }

    private void validateFixedLeaveQuota(LeaveType leaveType, int actualDays) {
        if (LeaveType.MARRIAGE.equals(leaveType)) {
            if (actualDays > 13) {
                throw new IllegalArgumentException("婚假最多13天（3天法定+10天晚婚）");
            }
        } else if (LeaveType.MATERNITY.equals(leaveType)) {
            if (actualDays > 158) {
                throw new IllegalArgumentException("产假最多158天");
            }
        } else if (LeaveType.PATERNITY.equals(leaveType)) {
            if (actualDays > 15) {
                throw new IllegalArgumentException("陪产假最多15天");
            }
        }
    }

    private void validateRequest(CreateLeaveRequest request) {
        if (request.getEmployeeId() == null) {
            throw new IllegalArgumentException("员工ID不能为空");
        }
        if (request.getLeaveType() == null) {
            throw new IllegalArgumentException("请假类型不能为空");
        }
        if (request.getStartDate() == null || request.getEndDate() == null) {
            throw new IllegalArgumentException("日期范围不能为空");
        }
        if (request.getStartDate().isAfter(request.getEndDate())) {
            throw new IllegalArgumentException("开始日期不能晚于结束日期");
        }
        if (request.getReason() == null || request.getReason().trim().isEmpty()) {
            throw new IllegalArgumentException("请假事由不能为空");
        }
    }

    public LeaveRecord approveLeave(Long leaveId, Long managerId, boolean approved, String comment) {
        Optional<LeaveRecord> recordOpt = leaveRecordRepository.findById(leaveId);
        if (recordOpt.isEmpty()) {
            throw new IllegalArgumentException("请假记录不存在");
        }
        
        LeaveRecord record = recordOpt.get();
        
        if (!LeaveStatus.PENDING.equals(record.getStatus())) {
            throw new IllegalArgumentException("该请假已被处理，无法再次操作");
        }
        
        Optional<Employee> employeeOpt = employeeService.getEmployee(record.getEmployeeId());
        if (employeeOpt.isPresent()) {
            Employee employee = employeeOpt.get();
            if (!managerId.equals(employee.getManagerId())) {
                throw new IllegalArgumentException("您不是该员工的上级，无法审批");
            }
        }
        
        if (approved) {
            if (LeaveType.ANNUAL.equals(record.getLeaveType())) {
                Optional<Employee> empOpt = employeeService.getEmployee(record.getEmployeeId());
                if (empOpt.isPresent()) {
                    Employee employee = empOpt.get();
                    annualLeaveService.deductAnnualLeave(employee, record.getActualDays());
                    employeeService.saveEmployee(employee);
                }
            }
            record.setStatus(LeaveStatus.APPROVED);
        } else {
            record.setStatus(LeaveStatus.REJECTED);
        }
        
        record.setApprovedBy(managerId);
        record.setApprovalComment(comment);
        
        return leaveRecordRepository.save(record);
    }

    public Optional<LeaveRecord> getLeaveRecord(Long id) {
        return leaveRecordRepository.findById(id);
    }

    public List<LeaveRecord> getLeaveRecordsByEmployee(Long employeeId) {
        return leaveRecordRepository.findByEmployeeId(employeeId);
    }

    public List<LeaveRecord> getAllLeaveRecords() {
        return leaveRecordRepository.findAll();
    }

    public List<LeaveRecord> getLeaveRecordsByDateRange(LocalDate startDate, LocalDate endDate) {
        return leaveRecordRepository.findByDateRange(startDate, endDate);
    }
}
