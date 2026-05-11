package com.chainrestaurant.shiftscheduler.service;

import com.chainrestaurant.shiftscheduler.entity.Employee;
import com.chainrestaurant.shiftscheduler.exception.BusinessException;
import com.chainrestaurant.shiftscheduler.exception.ResourceNotFoundException;
import com.chainrestaurant.shiftscheduler.repository.EmployeeRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
@Transactional
public class EmployeeService {
    
    @Autowired
    private EmployeeRepository employeeRepository;
    
    public Employee createEmployee(Employee employee) {
        if (employeeRepository.existsByEmployeeNo(employee.getEmployeeNo())) {
            throw new BusinessException("员工工号已存在: " + employee.getEmployeeNo());
        }
        return employeeRepository.save(employee);
    }
    
    public Employee updateEmployee(Long id, Employee employeeDetails) {
        Employee employee = getEmployeeById(id);
        
        if (!employee.getEmployeeNo().equals(employeeDetails.getEmployeeNo())) {
            if (employeeRepository.existsByEmployeeNo(employeeDetails.getEmployeeNo())) {
                throw new BusinessException("员工工号已存在: " + employeeDetails.getEmployeeNo());
            }
        }
        
        employee.setName(employeeDetails.getName());
        employee.setEmployeeNo(employeeDetails.getEmployeeNo());
        employee.setStoreName(employeeDetails.getStoreName());
        
        return employeeRepository.save(employee);
    }
    
    public void deleteEmployee(Long id) {
        if (!employeeRepository.existsById(id)) {
            throw new ResourceNotFoundException("员工不存在: " + id);
        }
        employeeRepository.deleteById(id);
    }
    
    public Employee getEmployeeById(Long id) {
        return employeeRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("员工不存在: " + id));
    }
    
    public List<Employee> getAllEmployees() {
        return employeeRepository.findAll();
    }
    
    public List<Employee> getEmployeesByStoreName(String storeName) {
        return employeeRepository.findByStoreName(storeName);
    }
    
    public Employee getEmployeeByNo(String employeeNo) {
        return employeeRepository.findByEmployeeNo(employeeNo)
                .orElseThrow(() -> new ResourceNotFoundException("员工不存在: " + employeeNo));
    }
}
