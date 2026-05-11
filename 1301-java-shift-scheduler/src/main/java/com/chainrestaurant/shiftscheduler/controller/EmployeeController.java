package com.chainrestaurant.shiftscheduler.controller;

import com.chainrestaurant.shiftscheduler.entity.Employee;
import com.chainrestaurant.shiftscheduler.service.EmployeeService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/employees")
public class EmployeeController {
    
    @Autowired
    private EmployeeService employeeService;
    
    @PostMapping
    public ResponseEntity<Employee> createEmployee(@RequestBody Employee employee) {
        Employee created = employeeService.createEmployee(employee);
        return ResponseEntity.ok(created);
    }
    
    @PutMapping("/{id}")
    public ResponseEntity<Employee> updateEmployee(@PathVariable Long id, @RequestBody Employee employee) {
        Employee updated = employeeService.updateEmployee(id, employee);
        return ResponseEntity.ok(updated);
    }
    
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteEmployee(@PathVariable Long id) {
        employeeService.deleteEmployee(id);
        return ResponseEntity.noContent().build();
    }
    
    @GetMapping("/{id}")
    public ResponseEntity<Employee> getEmployeeById(@PathVariable Long id) {
        Employee employee = employeeService.getEmployeeById(id);
        return ResponseEntity.ok(employee);
    }
    
    @GetMapping
    public ResponseEntity<List<Employee>> getAllEmployees(@RequestParam(required = false) String storeName) {
        if (storeName != null && !storeName.isEmpty()) {
            return ResponseEntity.ok(employeeService.getEmployeesByStoreName(storeName));
        }
        return ResponseEntity.ok(employeeService.getAllEmployees());
    }
    
    @GetMapping("/by-no/{employeeNo}")
    public ResponseEntity<Employee> getEmployeeByNo(@PathVariable String employeeNo) {
        Employee employee = employeeService.getEmployeeByNo(employeeNo);
        return ResponseEntity.ok(employee);
    }
}
