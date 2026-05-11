package com.company.expense.service;

import com.company.expense.dto.EmployeeRequest;
import com.company.expense.dto.EmployeeResponse;
import com.company.expense.entity.Employee;
import com.company.expense.exception.BusinessException;
import com.company.expense.exception.ResourceNotFoundException;
import com.company.expense.repository.EmployeeRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class EmployeeService {

    private final EmployeeRepository employeeRepository;

    @Transactional
    public EmployeeResponse createEmployee(EmployeeRequest request) {
        log.info("创建员工: {}", request.getEmployeeId());

        if (employeeRepository.findByEmployeeId(request.getEmployeeId()).isPresent()) {
            throw new BusinessException("员工编号已存在: " + request.getEmployeeId());
        }

        Employee employee = new Employee();
        employee.setEmployeeId(request.getEmployeeId());
        employee.setName(request.getName());
        employee.setDepartment(request.getDepartment());
        employee.setIsManager(request.getIsManager() != null ? request.getIsManager() : false);
        employee.setIsGeneralManager(request.getIsGeneralManager() != null ? request.getIsGeneralManager() : false);

        if (request.getSupervisorEmployeeId() != null && !request.getSupervisorEmployeeId().isEmpty()) {
            Employee supervisor = employeeRepository.findByEmployeeId(request.getSupervisorEmployeeId())
                    .orElseThrow(() -> new BusinessException("直属主管不存在: " + request.getSupervisorEmployeeId()));
            employee.setSupervisor(supervisor);
        }

        employee = employeeRepository.save(employee);
        log.info("员工创建成功: {}", employee.getEmployeeId());

        return EmployeeResponse.fromEntity(employee);
    }

    @Transactional
    public EmployeeResponse updateEmployee(String employeeId, EmployeeRequest request) {
        log.info("更新员工: {}", employeeId);

        Employee employee = employeeRepository.findByEmployeeId(employeeId)
                .orElseThrow(() -> new ResourceNotFoundException("员工不存在: " + employeeId));

        if (!employeeId.equals(request.getEmployeeId())) {
            if (employeeRepository.findByEmployeeId(request.getEmployeeId()).isPresent()) {
                throw new BusinessException("员工编号已存在: " + request.getEmployeeId());
            }
            employee.setEmployeeId(request.getEmployeeId());
        }

        employee.setName(request.getName());
        employee.setDepartment(request.getDepartment());
        employee.setIsManager(request.getIsManager() != null ? request.getIsManager() : false);
        employee.setIsGeneralManager(request.getIsGeneralManager() != null ? request.getIsGeneralManager() : false);

        if (request.getSupervisorEmployeeId() != null) {
            if (request.getSupervisorEmployeeId().isEmpty()) {
                employee.setSupervisor(null);
            } else {
                Employee supervisor = employeeRepository.findByEmployeeId(request.getSupervisorEmployeeId())
                        .orElseThrow(() -> new BusinessException("直属主管不存在: " + request.getSupervisorEmployeeId()));
                
                if (supervisor.getEmployeeId().equals(employee.getEmployeeId())) {
                    throw new BusinessException("直属主管不能是员工本人");
                }
                
                employee.setSupervisor(supervisor);
            }
        }

        employee = employeeRepository.save(employee);
        log.info("员工更新成功: {}", employee.getEmployeeId());

        return EmployeeResponse.fromEntity(employee);
    }

    public EmployeeResponse getEmployee(String employeeId) {
        log.debug("查询员工: {}", employeeId);

        Employee employee = employeeRepository.findByEmployeeId(employeeId)
                .orElseThrow(() -> new ResourceNotFoundException("员工不存在: " + employeeId));

        return EmployeeResponse.fromEntity(employee);
    }

    public EmployeeResponse getEmployeeById(Long id) {
        log.debug("查询员工 ID: {}", id);

        Employee employee = employeeRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("员工不存在，ID: " + id));

        return EmployeeResponse.fromEntity(employee);
    }

    public List<EmployeeResponse> getAllEmployees() {
        log.debug("查询所有员工");

        return employeeRepository.findAll().stream()
                .map(EmployeeResponse::fromEntity)
                .collect(Collectors.toList());
    }

    @Transactional
    public void deleteEmployee(String employeeId) {
        log.info("删除员工: {}", employeeId);

        Employee employee = employeeRepository.findByEmployeeId(employeeId)
                .orElseThrow(() -> new ResourceNotFoundException("员工不存在: " + employeeId));

        employeeRepository.delete(employee);
        log.info("员工删除成功: {}", employeeId);
    }

    public Employee getEmployeeEntity(String employeeId) {
        return employeeRepository.findByEmployeeId(employeeId)
                .orElseThrow(() -> new ResourceNotFoundException("员工不存在: " + employeeId));
    }
}
