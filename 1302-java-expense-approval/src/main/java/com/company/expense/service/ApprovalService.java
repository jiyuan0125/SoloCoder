package com.company.expense.service;

import com.company.expense.entity.Employee;
import com.company.expense.enums.ApprovalLevel;
import com.company.expense.repository.EmployeeRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

@Slf4j
@Service
@RequiredArgsConstructor
public class ApprovalService {

    private final EmployeeRepository employeeRepository;

    public Employee determineApprover(Employee applicant, ApprovalLevel level) {
        log.debug("为员工 {} 查找 {} 级别的审批人", 
                applicant.getEmployeeId(), level.getDescription());

        Employee approver = null;

        switch (level) {
            case SUPERVISOR:
                approver = findSupervisor(applicant);
                break;
            case MANAGER:
                approver = findDepartmentManager(applicant);
                break;
            case GENERAL_MANAGER:
                approver = findGeneralManager();
                break;
        }

        if (approver == null) {
            throw new RuntimeException("无法找到 " + level.getDescription() + " 级别的审批人");
        }

        if (approver.getEmployeeId().equals(applicant.getEmployeeId())) {
            log.warn("审批人 {} 与申请人 {} 相同，需要查找上一级审批人", 
                    approver.getEmployeeId(), applicant.getEmployeeId());
            return findNextHigherApprover(applicant, level);
        }

        log.debug("找到审批人: {} ({})", approver.getName(), approver.getEmployeeId());
        return approver;
    }

    private Employee findSupervisor(Employee employee) {
        if (employee.getSupervisor() != null) {
            return employee.getSupervisor();
        }
        return findDepartmentManager(employee);
    }

    private Employee findDepartmentManager(Employee employee) {
        return employeeRepository.findManagerByDepartment(employee.getDepartment())
                .orElse(null);
    }

    private Employee findGeneralManager() {
        return employeeRepository.findGeneralManager()
                .orElse(null);
    }

    private Employee findNextHigherApprover(Employee applicant, ApprovalLevel currentLevel) {
        log.debug("查找 {} 之上的审批人", currentLevel.getDescription());

        switch (currentLevel) {
            case SUPERVISOR:
                return determineApprover(applicant, ApprovalLevel.MANAGER);
            case MANAGER:
                return determineApprover(applicant, ApprovalLevel.GENERAL_MANAGER);
            case GENERAL_MANAGER:
                throw new RuntimeException("总经理报销需要特殊处理，没有更高级别审批人");
            default:
                throw new RuntimeException("未知审批级别: " + currentLevel);
        }
    }

    public ApprovalLevel getNextApprovalLevel(ApprovalLevel currentLevel) {
        switch (currentLevel) {
            case SUPERVISOR:
                return ApprovalLevel.MANAGER;
            case MANAGER:
                return ApprovalLevel.GENERAL_MANAGER;
            case GENERAL_MANAGER:
                return null;
            default:
                throw new RuntimeException("未知审批级别: " + currentLevel);
        }
    }

    public boolean hasMoreApprovalLevels(ApprovalLevel currentLevel) {
        return currentLevel != ApprovalLevel.GENERAL_MANAGER;
    }
}
