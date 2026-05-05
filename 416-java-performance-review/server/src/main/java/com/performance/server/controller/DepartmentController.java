package com.performance.server.controller;

import com.performance.common.constant.ErrorCode;
import com.performance.common.dto.ApiResponse;
import com.performance.common.dto.DepartmentDTO;
import com.performance.server.service.DepartmentService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/departments")
public class DepartmentController {

    @Autowired
    private DepartmentService departmentService;

    @PostMapping
    public ApiResponse<DepartmentDTO> createDepartment(@RequestBody DepartmentDTO dto) {
        DepartmentDTO created = departmentService.createDepartment(dto);
        return ApiResponse.success(created);
    }

    @GetMapping("/{id}")
    public ApiResponse<DepartmentDTO> getDepartmentById(@PathVariable Long id) {
        Optional<DepartmentDTO> deptOpt = departmentService.getDepartmentById(id);
        if (deptOpt.isPresent()) {
            return ApiResponse.success(deptOpt.get());
        }
        return ApiResponse.error(ErrorCode.NOT_FOUND);
    }

    @GetMapping
    public ApiResponse<List<DepartmentDTO>> getAllDepartments() {
        List<DepartmentDTO> departments = departmentService.getAllDepartments();
        return ApiResponse.success(departments);
    }
}