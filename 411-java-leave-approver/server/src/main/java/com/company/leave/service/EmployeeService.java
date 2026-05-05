package com.company.leave.service;

import com.company.leave.entity.Employee;
import com.company.leave.repository.EmployeeRepository;
import org.springframework.stereotype.Service;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;

@Service
public class EmployeeService {
    private final EmployeeRepository employeeRepository;
    private final AnnualLeaveService annualLeaveService;

    public EmployeeService(EmployeeRepository employeeRepository, AnnualLeaveService annualLeaveService) {
        this.employeeRepository = employeeRepository;
        this.annualLeaveService = annualLeaveService;
    }

    public Employee createEmployee(String name, Long managerId, LocalDate joinDate) {
        Employee employee = new Employee();
        employee.setName(name);
        employee.setManagerId(managerId);
        employee.setJoinDate(joinDate);
        
        int yearsOfService = annualLeaveService.calculateYearsOfService(joinDate, LocalDate.now());
        int quota = annualLeaveService.calculateAnnualLeaveQuota(yearsOfService);
        employee.setAnnualLeaveQuota(quota);
        employee.setAnnualLeaveRemaining(quota);
        employee.setCarriedOverLeave(0);
        
        return employeeRepository.save(employee);
    }

    public Optional<Employee> getEmployee(Long id) {
        return employeeRepository.findById(id);
    }

    public List<Employee> getAllEmployees() {
        return employeeRepository.findAll();
    }

    public List<Employee> getTeamMembers(Long managerId) {
        return employeeRepository.findByManagerId(managerId);
    }

    public Employee saveEmployee(Employee employee) {
        return employeeRepository.save(employee);
    }

    public boolean exists(Long id) {
        return employeeRepository.existsById(id);
    }
}
