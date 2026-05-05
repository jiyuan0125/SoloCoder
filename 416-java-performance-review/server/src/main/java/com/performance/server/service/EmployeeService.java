package com.performance.server.service;

import com.performance.common.dto.EmployeeDTO;
import com.performance.server.entity.Department;
import com.performance.server.entity.Employee;
import com.performance.server.repository.DepartmentRepository;
import com.performance.server.repository.EmployeeRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

@Service
public class EmployeeService {

    @Autowired
    private EmployeeRepository employeeRepository;

    @Autowired
    private DepartmentRepository departmentRepository;

    public EmployeeDTO createEmployee(EmployeeDTO dto) {
        Employee employee = new Employee();
        employee.setName(dto.getName());
        employee.setEmployeeNo(dto.getEmployeeNo());
        employee.setDepartmentId(dto.getDepartmentId());
        employee.setManagerId(dto.getManagerId());
        employee.setRole(dto.getRole());
        
        Employee saved = employeeRepository.save(employee);
        return toDTO(saved);
    }

    public Optional<EmployeeDTO> getEmployeeById(Long id) {
        Optional<Employee> employeeOpt = employeeRepository.findById(id);
        if (employeeOpt.isPresent()) {
            return Optional.of(toDTO(employeeOpt.get()));
        }
        return Optional.empty();
    }

    public List<EmployeeDTO> getAllEmployees() {
        List<Employee> employees = employeeRepository.findAll();
        List<EmployeeDTO> dtos = new ArrayList<>();
        for (Employee employee : employees) {
            dtos.add(toDTO(employee));
        }
        return dtos;
    }

    public List<EmployeeDTO> getEmployeesByDepartmentId(Long departmentId) {
        List<Employee> employees = employeeRepository.findByDepartmentId(departmentId);
        List<EmployeeDTO> dtos = new ArrayList<>();
        for (Employee employee : employees) {
            dtos.add(toDTO(employee));
        }
        return dtos;
    }

    public List<EmployeeDTO> getEmployeesByManagerId(Long managerId) {
        List<Employee> employees = employeeRepository.findByManagerId(managerId);
        List<EmployeeDTO> dtos = new ArrayList<>();
        for (Employee employee : employees) {
            dtos.add(toDTO(employee));
        }
        return dtos;
    }

    public Optional<EmployeeDTO> updateEmployeeDepartment(Long employeeId, Long newDepartmentId, Long newManagerId) {
        Optional<Employee> employeeOpt = employeeRepository.findById(employeeId);
        if (employeeOpt.isPresent()) {
            Employee employee = employeeOpt.get();
            employee.setDepartmentId(newDepartmentId);
            employee.setManagerId(newManagerId);
            Employee saved = employeeRepository.save(employee);
            return Optional.of(toDTO(saved));
        }
        return Optional.empty();
    }

    private EmployeeDTO toDTO(Employee employee) {
        EmployeeDTO dto = new EmployeeDTO();
        dto.setId(employee.getId());
        dto.setName(employee.getName());
        dto.setEmployeeNo(employee.getEmployeeNo());
        dto.setDepartmentId(employee.getDepartmentId());
        dto.setManagerId(employee.getManagerId());
        dto.setRole(employee.getRole());
        
        if (employee.getDepartmentId() != null) {
            Optional<Department> deptOpt = departmentRepository.findById(employee.getDepartmentId());
            if (deptOpt.isPresent()) {
                dto.setDepartmentName(deptOpt.get().getName());
            }
        }
        
        return dto;
    }
}