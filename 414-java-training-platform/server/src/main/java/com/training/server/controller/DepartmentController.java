package com.training.server.controller;

import com.training.common.dto.request.CreateDepartmentRequest;
import com.training.common.response.ApiResponse;
import com.training.common.dto.response.DepartmentDTO;
import com.training.common.enums.ErrorCode;
import com.training.server.service.DepartmentService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/api/departments")
public class DepartmentController {

    @Autowired
    private DepartmentService departmentService;

    @PostMapping
    public ApiResponse<DepartmentDTO> createDepartment(@RequestBody CreateDepartmentRequest request) {
        try {
            DepartmentDTO department = departmentService.createDepartment(request);
            return ApiResponse.success(department);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM, e.getMessage());
        }
    }

    @GetMapping("/{id}")
    public ApiResponse<DepartmentDTO> getDepartmentById(@PathVariable String id) {
        try {
            DepartmentDTO department = departmentService.getDepartmentById(id);
            return ApiResponse.success(department);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.DEPARTMENT_NOT_FOUND);
        }
    }

    @GetMapping
    public ApiResponse<List<DepartmentDTO>> getAllDepartments() {
        List<DepartmentDTO> departments = departmentService.getAllDepartments();
        return ApiResponse.success(departments);
    }

    @PutMapping("/{id}")
    public ApiResponse<DepartmentDTO> updateDepartment(
            @PathVariable String id, 
            @RequestBody CreateDepartmentRequest request) {
        try {
            DepartmentDTO department = departmentService.updateDepartment(id, request);
            return ApiResponse.success(department);
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.INVALID_PARAM, e.getMessage());
        }
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteDepartment(@PathVariable String id) {
        try {
            departmentService.deleteDepartment(id);
            return ApiResponse.success();
        } catch (RuntimeException e) {
            return ApiResponse.error(ErrorCode.DEPARTMENT_NOT_FOUND);
        }
    }
}
