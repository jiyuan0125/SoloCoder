package com.training.server.controller;

import com.training.common.dto.request.CreateEmployeeRequest;
import com.training.common.response.ApiResponse;
import com.training.common.dto.response.EmployeeDTO;
import com.training.common.enums.ErrorCode;
import com.training.server.service.EmployeeService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/employees")
public class EmployeeController {

    @Autowired
    private EmployeeService employeeService;

    @PostMapping
    public ApiResponse<EmployeeDTO> createEmployee(@RequestBody CreateEmployeeRequest request) {
        try {
            EmployeeDTO employee = employeeService.createEmployee(request);
            return ApiResponse.success(employee);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM, e.getMessage());
        }
    }

    @GetMapping("/{id}")
    public ApiResponse<EmployeeDTO> getEmployeeById(@PathVariable String id) {
        try {
            EmployeeDTO employee = employeeService.getEmployeeById(id);
            return ApiResponse.success(employee);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.EMPLOYEE_NOT_FOUND);
        }
    }

    @GetMapping
    public ApiResponse<List<EmployeeDTO>> getAllEmployees() {
        List<EmployeeDTO> employees = employeeService.getAllEmployees();
        return ApiResponse.success(employees);
    }

    @GetMapping(params = "departmentId")
    public ApiResponse<List<EmployeeDTO>> getEmployeesByDepartment(@RequestParam String departmentId) {
        try {
            List<EmployeeDTO> employees = employeeService.getEmployeesByDepartment(departmentId);
            return ApiResponse.success(employees);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.DEPARTMENT_NOT_FOUND);
        }
    }

    @PutMapping("/{id}")
    public ApiResponse<EmployeeDTO> updateEmployee(
            @PathVariable String id, 
            @RequestBody CreateEmployeeRequest request) {
        try {
            EmployeeDTO employee = employeeService.updateEmployee(id, request);
            return ApiResponse.success(employee);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM, e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteEmployee(@PathVariable String id) {
        try {
            employeeService.deleteEmployee(id);
            return ApiResponse.success();
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.EMPLOYEE_NOT_FOUND);
        }
    }
}
