package com.safety.inspection.controller;

import com.safety.inspection.common.Result;
import com.safety.inspection.dto.DepartmentDTO;
import com.safety.inspection.entity.Department;
import com.safety.inspection.service.DepartmentService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.security.access.prepost.PreAuthorize;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/departments")
@RequiredArgsConstructor
public class DepartmentController {

    private final DepartmentService departmentService;

    @GetMapping("/list")
    public Result<List<Department>> getAllDepartments(@RequestParam(required = false) String deptName) {
        return Result.success(departmentService.getAllDepartments(deptName));
    }

    @GetMapping("/{id}")
    public Result<Department> getDepartmentById(@PathVariable Long id) {
        return Result.success(departmentService.getDepartmentById(id));
    }

    @PostMapping
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> createDepartment(@RequestBody @Valid DepartmentDTO dto) {
        departmentService.createDepartment(dto);
        return Result.success();
    }

    @PutMapping
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> updateDepartment(@RequestBody @Valid DepartmentDTO dto) {
        departmentService.updateDepartment(dto);
        return Result.success();
    }

    @DeleteMapping("/{id}")
    @PreAuthorize("hasRole('ADMIN')")
    public Result<Void> deleteDepartment(@PathVariable Long id) {
        departmentService.deleteDepartment(id);
        return Result.success();
    }
}
