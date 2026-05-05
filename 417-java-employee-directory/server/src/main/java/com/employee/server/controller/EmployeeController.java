package com.employee.server.controller;

import com.employee.common.dto.*;
import com.employee.common.response.ApiResponse;
import com.employee.server.service.EmployeeService;
import com.employee.server.service.OperationLogService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/employees")
public class EmployeeController {

    private final EmployeeService employeeService;
    private final OperationLogService operationLogService;

    @Autowired
    public EmployeeController(EmployeeService employeeService,
                              OperationLogService operationLogService) {
        this.employeeService = employeeService;
        this.operationLogService = operationLogService;
    }

    @PostMapping
    public ApiResponse<EmployeeDTO> createEmployee(@RequestBody EmployeeCreateRequest request) {
        EmployeeDTO employee = employeeService.createEmployee(request);
        return ApiResponse.success(employee);
    }

    @PutMapping("/{employeeId}")
    public ApiResponse<EmployeeDTO> updateEmployee(@PathVariable String employeeId,
                                                    @RequestBody EmployeeUpdateRequest request) {
        EmployeeDTO employee = employeeService.updateEmployee(employeeId, request);
        return ApiResponse.success(employee);
    }

    @PutMapping("/self")
    public ApiResponse<EmployeeDTO> updateSelfInfo(@RequestBody EmployeeSelfUpdateRequest request) {
        EmployeeDTO employee = employeeService.updateSelfInfo(request);
        return ApiResponse.success(employee);
    }

    @GetMapping("/{employeeId}")
    public ApiResponse<EmployeeDTO> getEmployeeById(@PathVariable String employeeId,
                                                      @RequestParam(required = false, defaultValue = "false") boolean includeResigned) {
        EmployeeDTO employee = employeeService.getEmployeeById(employeeId, includeResigned);
        return ApiResponse.success(employee);
    }

    @PostMapping("/search")
    public ApiResponse<List<EmployeeDTO>> searchEmployees(@RequestBody SearchRequest request) {
        List<EmployeeDTO> employees = employeeService.searchEmployees(request);
        return ApiResponse.success(employees);
    }

    @GetMapping
    public ApiResponse<List<EmployeeDTO>> listEmployees(@RequestParam(required = false, defaultValue = "false") boolean includeResigned) {
        SearchRequest request = new SearchRequest();
        request.setIncludeResigned(includeResigned);
        List<EmployeeDTO> employees = employeeService.searchEmployees(request);
        return ApiResponse.success(employees);
    }

    @PostMapping("/batch-import")
    public ApiResponse<BatchImportResultDTO> batchImport(@RequestBody List<EmployeeCreateRequest> employees) {
        BatchImportResultDTO result = employeeService.batchImport(employees);
        return ApiResponse.success(result);
    }

    @GetMapping("/stats/departments")
    public ApiResponse<List<DepartmentStatsDTO>> getDepartmentStats() {
        List<DepartmentStatsDTO> stats = employeeService.getDepartmentStats();
        return ApiResponse.success(stats);
    }

    @GetMapping("/organization-tree")
    public ApiResponse<OrgTreeNodeDTO> getOrganizationTree() {
        OrgTreeNodeDTO tree = employeeService.getOrganizationTree();
        return ApiResponse.success(tree);
    }

    @GetMapping("/logs")
    public ApiResponse<List<OperationLogDTO>> getAllLogs() {
        List<OperationLogDTO> logs = operationLogService.getAllLogs();
        return ApiResponse.success(logs);
    }

    @GetMapping("/{employeeId}/logs")
    public ApiResponse<List<OperationLogDTO>> getEmployeeLogs(@PathVariable String employeeId) {
        List<OperationLogDTO> logs = operationLogService.getLogsByEmployee(employeeId);
        return ApiResponse.success(logs);
    }
}
