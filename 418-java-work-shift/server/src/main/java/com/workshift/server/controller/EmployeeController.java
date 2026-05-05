package com.workshift.server.controller;

import com.workshift.common.dto.EmployeeDTO;
import com.workshift.common.enums.ErrorCode;
import com.workshift.common.response.ApiResponse;
import com.workshift.server.service.EmployeeService;
import java.util.List;
import java.util.Optional;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/employees")
public class EmployeeController {

    private final EmployeeService employeeService;

    public EmployeeController(EmployeeService employeeService) {
        this.employeeService = employeeService;
    }

    @PostMapping
    public ApiResponse<EmployeeDTO> createEmployee(@RequestBody EmployeeDTO employee) {
        if (employee.getName() == null || employee.getName().trim().isEmpty()) {
            return ApiResponse.error(ErrorCode.INVALID_PARAMETER.getCode(), "员工名称不能为空");
        }
        if (employee.getHourlyWage() == null) {
            return ApiResponse.error(ErrorCode.INVALID_PARAMETER.getCode(), "时薪不能为空");
        }
        EmployeeDTO created = employeeService.createEmployee(employee);
        return ApiResponse.success(created);
    }

    @PutMapping("/{id}")
    public ApiResponse<Void> updateEmployee(@PathVariable String id, @RequestBody EmployeeDTO employee) {
        Optional<ErrorCode> result = employeeService.updateEmployee(id, employee);
        if (result.isPresent()) {
            return ApiResponse.error(result.get().getCode(), result.get().getMessage());
        }
        return ApiResponse.success();
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteEmployee(@PathVariable String id) {
        Optional<ErrorCode> result = employeeService.deleteEmployee(id);
        if (result.isPresent()) {
            return ApiResponse.error(result.get().getCode(), result.get().getMessage());
        }
        return ApiResponse.success();
    }

    @GetMapping("/{id}")
    public ApiResponse<EmployeeDTO> getEmployee(@PathVariable String id) {
        EmployeeDTO employee = employeeService.getEmployee(id);
        if (employee == null) {
            return ApiResponse.error(ErrorCode.EMPLOYEE_NOT_FOUND.getCode(), ErrorCode.EMPLOYEE_NOT_FOUND.getMessage());
        }
        return ApiResponse.success(employee);
    }

    @GetMapping
    public ApiResponse<List<EmployeeDTO>> getAllEmployees() {
        List<EmployeeDTO> employees = employeeService.getAllEmployees();
        return ApiResponse.success(employees);
    }
}