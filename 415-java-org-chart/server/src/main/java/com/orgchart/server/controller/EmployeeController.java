package com.orgchart.server.controller;

import com.orgchart.common.dto.ApiResponse;
import com.orgchart.common.dto.EmployeeDTO;
import com.orgchart.common.dto.ReportingLineDTO;
import com.orgchart.common.dto.TransferHistoryDTO;
import com.orgchart.common.dto.request.CreateEmployeeRequest;
import com.orgchart.common.dto.request.TransferEmployeeRequest;
import com.orgchart.common.dto.request.UpdateEmployeeRequest;
import com.orgchart.server.service.EmployeeService;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/employees")
public class EmployeeController {

    private final EmployeeService employeeService;

    public EmployeeController(EmployeeService employeeService) {
        this.employeeService = employeeService;
    }

    @PostMapping
    public ApiResponse<EmployeeDTO> createEmployee(@RequestBody CreateEmployeeRequest request) {
        EmployeeDTO employee = employeeService.createEmployee(request);
        return ApiResponse.success(employee);
    }

    @PutMapping("/{id}")
    public ApiResponse<EmployeeDTO> updateEmployee(@PathVariable String id, 
                                                     @RequestBody UpdateEmployeeRequest request) {
        EmployeeDTO employee = employeeService.updateEmployee(id, request);
        return ApiResponse.success(employee);
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteEmployee(@PathVariable String id) {
        employeeService.deleteEmployee(id);
        return ApiResponse.success();
    }

    @GetMapping("/{id}")
    public ApiResponse<EmployeeDTO> getEmployeeById(@PathVariable String id) {
        EmployeeDTO employee = employeeService.getEmployeeById(id);
        return ApiResponse.success(employee);
    }

    @GetMapping
    public ApiResponse<List<EmployeeDTO>> getAllEmployees() {
        List<EmployeeDTO> employees = employeeService.getAllEmployees();
        return ApiResponse.success(employees);
    }

    @GetMapping("/department/{departmentId}")
    public ApiResponse<List<EmployeeDTO>> getEmployeesByDepartment(
            @PathVariable String departmentId,
            @RequestParam(defaultValue = "false") boolean includeSubDepartments) {
        List<EmployeeDTO> employees = employeeService.getEmployeesByDepartment(departmentId, includeSubDepartments);
        return ApiResponse.success(employees);
    }

    @PostMapping("/{id}/transfer")
    public ApiResponse<EmployeeDTO> transferEmployee(@PathVariable String id, 
                                                       @RequestBody TransferEmployeeRequest request) {
        EmployeeDTO employee = employeeService.transferEmployee(id, request);
        return ApiResponse.success(employee);
    }

    @GetMapping("/{id}/reporting-line")
    public ApiResponse<ReportingLineDTO> getReportingLine(@PathVariable String id) {
        ReportingLineDTO reportingLine = employeeService.getReportingLine(id);
        return ApiResponse.success(reportingLine);
    }

    @GetMapping("/{id}/transfer-history")
    public ApiResponse<List<TransferHistoryDTO>> getTransferHistory(@PathVariable String id) {
        List<TransferHistoryDTO> history = employeeService.getTransferHistory(id);
        return ApiResponse.success(history);
    }
}
