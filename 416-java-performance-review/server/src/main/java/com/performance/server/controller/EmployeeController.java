package com.performance.server.controller;

import com.performance.common.constant.ErrorCode;
import com.performance.common.dto.ApiResponse;
import com.performance.common.dto.EmployeeDTO;
import com.performance.server.service.EmployeeService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/employees")
public class EmployeeController {

    @Autowired
    private EmployeeService employeeService;

    @PostMapping
    public ApiResponse<EmployeeDTO> createEmployee(@RequestBody EmployeeDTO dto) {
        EmployeeDTO created = employeeService.createEmployee(dto);
        return ApiResponse.success(created);
    }

    @GetMapping("/{id}")
    public ApiResponse<EmployeeDTO> getEmployeeById(@PathVariable Long id) {
        Optional<EmployeeDTO> employeeOpt = employeeService.getEmployeeById(id);
        if (employeeOpt.isPresent()) {
            return ApiResponse.success(employeeOpt.get());
        }
        return ApiResponse.error(ErrorCode.NOT_FOUND);
    }

    @GetMapping
    public ApiResponse<List<EmployeeDTO>> getAllEmployees() {
        List<EmployeeDTO> employees = employeeService.getAllEmployees();
        return ApiResponse.success(employees);
    }

    @GetMapping("/department/{departmentId}")
    public ApiResponse<List<EmployeeDTO>> getEmployeesByDepartmentId(@PathVariable Long departmentId) {
        List<EmployeeDTO> employees = employeeService.getEmployeesByDepartmentId(departmentId);
        return ApiResponse.success(employees);
    }

    @PutMapping("/{employeeId}/transfer")
    public ApiResponse<EmployeeDTO> transferEmployee(
            @PathVariable Long employeeId,
            @RequestParam Long newDepartmentId,
            @RequestParam Long newManagerId) {
        Optional<EmployeeDTO> updated = employeeService.updateEmployeeDepartment(employeeId, newDepartmentId, newManagerId);
        if (updated.isPresent()) {
            return ApiResponse.success(updated.get());
        }
        return ApiResponse.error(ErrorCode.NOT_FOUND);
    }
}