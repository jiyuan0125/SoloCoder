package com.training.server.service;

import com.training.common.dto.request.CreateEmployeeRequest;
import com.training.common.dto.response.DepartmentDTO;
import com.training.common.dto.response.EmployeeDTO;
import com.training.common.enums.ErrorCode;
import com.training.server.entity.Department;
import com.training.server.entity.Employee;
import com.training.server.repository.DepartmentRepository;
import com.training.server.repository.EmployeeRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Service
public class EmployeeService {

    @Autowired
    private EmployeeRepository employeeRepository;

    @Autowired
    private DepartmentRepository departmentRepository;

    public EmployeeDTO createEmployee(CreateEmployeeRequest request) {
        if (request.getDepartmentId() != null && !departmentRepository.existsById(request.getDepartmentId())) {
            throw new RuntimeException(ErrorCode.DEPARTMENT_NOT_FOUND.getMessage());
        }

        Employee employee = new Employee();
        employee.setId(UUID.randomUUID().toString());
        employee.setName(request.getName());
        employee.setDepartmentId(request.getDepartmentId());
        employee.setEmail(request.getEmail());
        employee.setSalary(request.getSalary());

        Employee saved = employeeRepository.save(employee);
        return toDTO(saved);
    }

    public EmployeeDTO getEmployeeById(String id) {
        Optional<Employee> employeeOpt = employeeRepository.findById(id);
        if (employeeOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.EMPLOYEE_NOT_FOUND.getMessage());
        }
        return toDTO(employeeOpt.get());
    }

    public List<EmployeeDTO> getAllEmployees() {
        List<Employee> employees = employeeRepository.findAll();
        List<EmployeeDTO> dtoList = new ArrayList<>();
        for (Employee employee : employees) {
            dtoList.add(toDTO(employee));
        }
        return dtoList;
    }

    public List<EmployeeDTO> getEmployeesByDepartment(String departmentId) {
        if (!departmentRepository.existsById(departmentId)) {
            throw new RuntimeException(ErrorCode.DEPARTMENT_NOT_FOUND.getMessage());
        }
        List<Employee> employees = employeeRepository.findByDepartmentId(departmentId);
        List<EmployeeDTO> dtoList = new ArrayList<>();
        for (Employee employee : employees) {
            dtoList.add(toDTO(employee));
        }
        return dtoList;
    }

    public EmployeeDTO updateEmployee(String id, CreateEmployeeRequest request) {
        Optional<Employee> employeeOpt = employeeRepository.findById(id);
        if (employeeOpt.isEmpty()) {
            throw new RuntimeException(ErrorCode.EMPLOYEE_NOT_FOUND.getMessage());
        }

        if (request.getDepartmentId() != null && !departmentRepository.existsById(request.getDepartmentId())) {
            throw new RuntimeException(ErrorCode.DEPARTMENT_NOT_FOUND.getMessage());
        }

        Employee employee = employeeOpt.get();
        if (request.getName() != null) {
            employee.setName(request.getName());
        }
        if (request.getDepartmentId() != null) {
            employee.setDepartmentId(request.getDepartmentId());
        }
        if (request.getEmail() != null) {
            employee.setEmail(request.getEmail());
        }
        if (request.getSalary() != null) {
            employee.setSalary(request.getSalary());
        }

        Employee saved = employeeRepository.save(employee);
        return toDTO(saved);
    }

    public void deleteEmployee(String id) {
        if (!employeeRepository.existsById(id)) {
            throw new RuntimeException(ErrorCode.EMPLOYEE_NOT_FOUND.getMessage());
        }
        employeeRepository.deleteById(id);
    }

    private EmployeeDTO toDTO(Employee employee) {
        EmployeeDTO dto = new EmployeeDTO();
        dto.setId(employee.getId());
        dto.setName(employee.getName());
        dto.setDepartmentId(employee.getDepartmentId());
        dto.setEmail(employee.getEmail());
        dto.setSalary(employee.getSalary());
        dto.setActive(employee.isActive());
        dto.setCreatedAt(employee.getCreatedAt());

        if (employee.getDepartmentId() != null) {
            Optional<Department> deptOpt = departmentRepository.findById(employee.getDepartmentId());
            if (deptOpt.isPresent()) {
                dto.setDepartmentName(deptOpt.get().getName());
            }
        }

        return dto;
    }
}
