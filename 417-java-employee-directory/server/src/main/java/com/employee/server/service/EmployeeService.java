package com.employee.server.service;

import com.employee.common.constant.EmployeeConstants;
import com.employee.common.constant.ErrorCode;
import com.employee.common.dto.*;
import com.employee.server.context.UserContext;
import com.employee.server.entity.Employee;
import com.employee.server.repository.EmployeeRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.*;
import java.util.stream.Collectors;

@Service
public class EmployeeService {
    private final EmployeeRepository employeeRepository;
    private final OperationLogService operationLogService;

    @Autowired
    public EmployeeService(EmployeeRepository employeeRepository,
                           OperationLogService operationLogService) {
        this.employeeRepository = employeeRepository;
        this.operationLogService = operationLogService;
    }

    public EmployeeDTO createEmployee(EmployeeCreateRequest request) {
        validateCreateRequest(request);

        if (employeeRepository.existsById(request.getEmployeeId())) {
            throw new BusinessException(ErrorCode.EMPLOYEE_ID_EXISTS);
        }

        validateEmailDomain(request.getEmail());
        validatePhoneUniqueness(request.getPhone(), request.getEmployeeId());

        Employee employee = new Employee();
        employee.setEmployeeId(request.getEmployeeId());
        employee.setName(request.getName());
        employee.setDepartment(request.getDepartment());
        employee.setPosition(request.getPosition());
        employee.setPhone(request.getPhone());
        employee.setEmail(request.getEmail());
        employee.setOfficeLocation(request.getOfficeLocation());
        employee.setResigned(false);

        employeeRepository.save(employee);

        return toDTO(employee);
    }

    public EmployeeDTO updateEmployee(String employeeId, EmployeeUpdateRequest request) {
        Employee employee = findEmployeeOrThrow(employeeId);
        String oldName = employee.getName();

        if (!UserContext.canModifyAllFields()) {
            throw new BusinessException(ErrorCode.PERMISSION_DENIED);
        }

        if (request.getName() != null && !request.getName().equals(employee.getName())) {
            operationLogService.logChange(employeeId, oldName, "姓名",
                    employee.getName(), request.getName());
            employee.setName(request.getName());
        }

        if (request.getDepartment() != null && !request.getDepartment().equals(employee.getDepartment())) {
            operationLogService.logChange(employeeId, oldName, "部门",
                    employee.getDepartment(), request.getDepartment());
            employee.setDepartment(request.getDepartment());
        }

        if (request.getPosition() != null && !request.getPosition().equals(employee.getPosition())) {
            operationLogService.logChange(employeeId, oldName, "职位",
                    employee.getPosition(), request.getPosition());
            employee.setPosition(request.getPosition());
        }

        if (request.getPhone() != null && !request.getPhone().equals(employee.getPhone())) {
            validatePhoneUniqueness(request.getPhone(), employeeId);
            operationLogService.logChange(employeeId, oldName, "手机号",
                    employee.getPhone(), request.getPhone());
            employee.setPhone(request.getPhone());
        }

        if (request.getEmail() != null && !request.getEmail().equals(employee.getEmail())) {
            validateEmailDomain(request.getEmail());
            operationLogService.logChange(employeeId, oldName, "邮箱",
                    employee.getEmail(), request.getEmail());
            employee.setEmail(request.getEmail());
        }

        if (request.getOfficeLocation() != null && !request.getOfficeLocation().equals(employee.getOfficeLocation())) {
            operationLogService.logChange(employeeId, oldName, "办公地点",
                    employee.getOfficeLocation(), request.getOfficeLocation());
            employee.setOfficeLocation(request.getOfficeLocation());
        }

        if (request.getResigned() != null && request.getResigned() != employee.isResigned()) {
            operationLogService.logChange(employeeId, oldName, "离职状态",
                    String.valueOf(employee.isResigned()), String.valueOf(request.getResigned()));
            employee.setResigned(request.getResigned());
        }

        employee.setUpdatedAt(LocalDateTime.now());
        employeeRepository.save(employee);

        return toDTO(employee);
    }

    public EmployeeDTO updateSelfInfo(EmployeeSelfUpdateRequest request) {
        String currentUserId = UserContext.getCurrentUserId();
        if (currentUserId == null) {
            throw new BusinessException(ErrorCode.PERMISSION_DENIED);
        }

        Employee employee = findEmployeeOrThrow(currentUserId);
        String oldName = employee.getName();

        if (request.getPhone() != null && !request.getPhone().equals(employee.getPhone())) {
            validatePhoneUniqueness(request.getPhone(), currentUserId);
            operationLogService.logChange(currentUserId, oldName, "手机号",
                    employee.getPhone(), request.getPhone());
            employee.setPhone(request.getPhone());
        }

        if (request.getEmail() != null && !request.getEmail().equals(employee.getEmail())) {
            validateEmailDomain(request.getEmail());
            operationLogService.logChange(currentUserId, oldName, "邮箱",
                    employee.getEmail(), request.getEmail());
            employee.setEmail(request.getEmail());
        }

        employee.setUpdatedAt(LocalDateTime.now());
        employeeRepository.save(employee);

        return toDTO(employee);
    }

    public EmployeeDTO getEmployeeById(String employeeId, boolean includeResigned) {
        Optional<Employee> optionalEmployee = employeeRepository.findById(employeeId);
        
        if (optionalEmployee.isEmpty()) {
            throw new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND);
        }
        
        Employee employee = optionalEmployee.get();
        
        if (!includeResigned && employee.isResigned() && !UserContext.canViewResigned()) {
            throw new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND);
        }
        
        return toDTO(employee);
    }

    public List<EmployeeDTO> searchEmployees(SearchRequest request) {
        boolean includeResigned = request.isIncludeResigned();
        if (includeResigned && !UserContext.canViewResigned()) {
            includeResigned = false;
        }

        List<Employee> employees;
        if (request.getKeyword() == null || request.getKeyword().isEmpty()) {
            if (includeResigned) {
                employees = employeeRepository.findAll();
            } else {
                employees = employeeRepository.findByResigned(false);
            }
        } else {
            employees = employeeRepository.searchByKeyword(request.getKeyword(), includeResigned);
        }

        return employees.stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
    }

    public List<DepartmentStatsDTO> getDepartmentStats() {
        Map<String, DepartmentStatsDTO> statsMap = new HashMap<>();

        for (Employee employee : employeeRepository.findAll()) {
            String dept = employee.getDepartment();
            DepartmentStatsDTO stats = statsMap.computeIfAbsent(dept, k -> {
                DepartmentStatsDTO s = new DepartmentStatsDTO();
                s.setDepartment(k);
                return s;
            });

            stats.setTotalCount(stats.getTotalCount() + 1);
            if (employee.isResigned()) {
                stats.setResignedCount(stats.getResignedCount() + 1);
            } else {
                stats.setActiveCount(stats.getActiveCount() + 1);
            }
        }

        return new ArrayList<>(statsMap.values());
    }

    public OrgTreeNodeDTO getOrganizationTree() {
        OrgTreeNodeDTO root = new OrgTreeNodeDTO();
        root.setName("公司");
        root.setType("company");

        Map<String, OrgTreeNodeDTO> deptMap = new HashMap<>();

        for (Employee employee : employeeRepository.findAll()) {
            if (employee.isResigned()) {
                continue;
            }
            String dept = employee.getDepartment();
            OrgTreeNodeDTO deptNode = deptMap.computeIfAbsent(dept, k -> {
                OrgTreeNodeDTO d = new OrgTreeNodeDTO();
                d.setName(k);
                d.setType("department");
                root.addChild(d);
                return d;
            });

            deptNode.setEmployeeCount(deptNode.getEmployeeCount() + 1);
        }

        int totalCount = employeeRepository.findByResigned(false).size();
        root.setEmployeeCount(totalCount);

        return root;
    }

    public BatchImportResultDTO batchImport(List<EmployeeCreateRequest> employees) {
        BatchImportResultDTO result = new BatchImportResultDTO();
        result.setTotalCount(employees.size());

        for (int i = 0; i < employees.size(); i++) {
            EmployeeCreateRequest request = employees.get(i);
            String employeeId = request.getEmployeeId();

            try {
                if (employeeId == null || employeeId.isEmpty()) {
                    throw new BusinessException(ErrorCode.PARAM_INVALID, "工号不能为空");
                }

                if (employeeRepository.existsById(employeeId)) {
                    throw new BusinessException(ErrorCode.EMPLOYEE_ID_EXISTS, "工号已存在");
                }

                validateCreateRequest(request);
                validateEmailDomain(request.getEmail());
                validatePhoneUniqueness(request.getPhone(), employeeId);

                Employee employee = new Employee();
                employee.setEmployeeId(employeeId);
                employee.setName(request.getName());
                employee.setDepartment(request.getDepartment());
                employee.setPosition(request.getPosition());
                employee.setPhone(request.getPhone());
                employee.setEmail(request.getEmail());
                employee.setOfficeLocation(request.getOfficeLocation());
                employee.setResigned(false);

                employeeRepository.save(employee);
                result.setSuccessCount(result.getSuccessCount() + 1);
            } catch (BusinessException e) {
                result.addError(i, employeeId, e.getMessage());
            } catch (Exception e) {
                result.addError(i, employeeId, e.getMessage());
            }
        }

        return result;
    }

    private Employee findEmployeeOrThrow(String employeeId) {
        return employeeRepository.findById(employeeId)
                .orElseThrow(() -> new BusinessException(ErrorCode.EMPLOYEE_NOT_FOUND));
    }

    private void validateCreateRequest(EmployeeCreateRequest request) {
        if (request.getEmployeeId() == null || request.getEmployeeId().isEmpty()) {
            throw new BusinessException(ErrorCode.PARAM_INVALID, "工号不能为空");
        }
        if (request.getName() == null || request.getName().isEmpty()) {
            throw new BusinessException(ErrorCode.PARAM_INVALID, "姓名不能为空");
        }
        if (request.getDepartment() == null || request.getDepartment().isEmpty()) {
            throw new BusinessException(ErrorCode.PARAM_INVALID, "部门不能为空");
        }
        if (request.getEmail() == null || request.getEmail().isEmpty()) {
            throw new BusinessException(ErrorCode.PARAM_INVALID, "邮箱不能为空");
        }
    }

    private void validateEmailDomain(String email) {
        if (email == null) {
            return;
        }
        String domain = getEmailDomain(email);
        if (!EmployeeConstants.COMPANY_EMAIL_DOMAIN.equals(domain)) {
            throw new BusinessException(ErrorCode.EMAIL_INVALID_DOMAIN);
        }
    }

    private String getEmailDomain(String email) {
        if (email == null || !email.contains("@")) {
            return "";
        }
        return email.substring(email.indexOf("@") + 1).toLowerCase();
    }

    private void validatePhoneUniqueness(String phone, String employeeId) {
        if (phone == null) {
            return;
        }
        if (employeeRepository.isPhoneUsedByOtherActiveEmployee(phone, employeeId)) {
            throw new BusinessException(ErrorCode.PHONE_ALREADY_USED);
        }
    }

    private EmployeeDTO toDTO(Employee employee) {
        EmployeeDTO dto = new EmployeeDTO();
        dto.setEmployeeId(employee.getEmployeeId());
        dto.setName(employee.getName());
        dto.setDepartment(employee.getDepartment());
        dto.setPosition(employee.getPosition());
        dto.setPhone(employee.getPhone());
        dto.setEmail(employee.getEmail());
        dto.setOfficeLocation(employee.getOfficeLocation());
        dto.setResigned(employee.isResigned());
        dto.setCreatedAt(employee.getCreatedAt());
        dto.setUpdatedAt(employee.getUpdatedAt());
        return dto;
    }

    public static class BusinessException extends RuntimeException {
        private final ErrorCode errorCode;

        public BusinessException(ErrorCode errorCode) {
            super(errorCode.getMessage());
            this.errorCode = errorCode;
        }

        public BusinessException(ErrorCode errorCode, String message) {
            super(message);
            this.errorCode = errorCode;
        }

        public ErrorCode getErrorCode() {
            return errorCode;
        }
    }
}
