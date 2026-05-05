package com.orgchart.server.controller;

import com.orgchart.common.dto.ApiResponse;
import com.orgchart.common.dto.DepartmentDTO;
import com.orgchart.common.dto.request.CreateDepartmentRequest;
import com.orgchart.common.dto.request.MergeDepartmentRequest;
import com.orgchart.common.dto.request.UpdateDepartmentRequest;
import com.orgchart.server.service.DepartmentService;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/departments")
public class DepartmentController {

    private final DepartmentService departmentService;

    public DepartmentController(DepartmentService departmentService) {
        this.departmentService = departmentService;
    }

    @PostMapping
    public ApiResponse<DepartmentDTO> createDepartment(@RequestBody CreateDepartmentRequest request) {
        DepartmentDTO department = departmentService.createDepartment(request);
        return ApiResponse.success(department);
    }

    @PutMapping("/{id}")
    public ApiResponse<DepartmentDTO> updateDepartment(@PathVariable String id, 
                                                         @RequestBody UpdateDepartmentRequest request) {
        DepartmentDTO department = departmentService.updateDepartment(id, request);
        return ApiResponse.success(department);
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteDepartment(@PathVariable String id) {
        departmentService.deleteDepartment(id);
        return ApiResponse.success();
    }

    @GetMapping("/{id}")
    public ApiResponse<DepartmentDTO> getDepartmentById(@PathVariable String id) {
        DepartmentDTO department = departmentService.getDepartmentById(id);
        return ApiResponse.success(department);
    }

    @GetMapping
    public ApiResponse<List<DepartmentDTO>> getAllDepartments() {
        List<DepartmentDTO> departments = departmentService.getAllDepartments();
        return ApiResponse.success(departments);
    }

    @GetMapping("/tree")
    public ApiResponse<List<DepartmentDTO>> getDepartmentTree() {
        List<DepartmentDTO> tree = departmentService.getDepartmentTree();
        return ApiResponse.success(tree);
    }

    @PostMapping("/{id}/merge")
    public ApiResponse<Void> mergeDepartment(@PathVariable String id, 
                                               @RequestBody MergeDepartmentRequest request) {
        departmentService.mergeDepartment(id, request);
        return ApiResponse.success();
    }
}
