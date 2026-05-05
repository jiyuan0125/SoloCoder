package com.company.leave.controller;

import com.company.leave.dto.ApiResponse;
import com.company.leave.dto.EmployeeDTO;
import com.company.leave.entity.Employee;
import com.company.leave.service.AnnualLeaveService;
import com.company.leave.service.EmployeeService;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDate;
import java.util.List;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/api/employees")
public class EmployeeController {
    private final EmployeeService employeeService;
    private final AnnualLeaveService annualLeaveService;

    public EmployeeController(EmployeeService employeeService, AnnualLeaveService annualLeaveService) {
        this.employeeService = employeeService;
        this.annualLeaveService = annualLeaveService;
    }

    @PostMapping
    public ApiResponse<EmployeeDTO> createEmployee(@RequestBody CreateEmployeeRequest request) {
        Employee employee = employeeService.createEmployee(
                request.getName(),
                request.getManagerId(),
                request.getJoinDate()
        );
        return ApiResponse.success(toDTO(employee));
    }

    @GetMapping("/{id}")
    public ApiResponse<EmployeeDTO> getEmployee(@PathVariable("id") Long id) {
        return employeeService.getEmployee(id)
                .map(emp -> ApiResponse.success(toDTO(emp)))
                .orElse(ApiResponse.error(com.company.leave.enums.ErrorCode.EMPLOYEE_NOT_FOUND));
    }

    @GetMapping
    public ApiResponse<List<EmployeeDTO>> getAllEmployees() {
        List<EmployeeDTO> dtos = employeeService.getAllEmployees().stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
        return ApiResponse.success(dtos);
    }

    @GetMapping("/manager/{managerId}/team")
    public ApiResponse<List<EmployeeDTO>> getTeamMembers(@PathVariable("managerId") Long managerId) {
        List<EmployeeDTO> dtos = employeeService.getTeamMembers(managerId).stream()
                .map(this::toDTO)
                .collect(Collectors.toList());
        return ApiResponse.success(dtos);
    }

    private EmployeeDTO toDTO(Employee employee) {
        EmployeeDTO dto = new EmployeeDTO();
        dto.setId(employee.getId());
        dto.setName(employee.getName());
        dto.setManagerId(employee.getManagerId());
        dto.setJoinDate(employee.getJoinDate());
        
        int yearsOfService = annualLeaveService.calculateYearsOfService(
                employee.getJoinDate(), LocalDate.now());
        dto.setYearsOfService(yearsOfService);
        
        dto.setAnnualLeaveQuota(employee.getAnnualLeaveQuota());
        dto.setAnnualLeaveRemaining(employee.getAnnualLeaveRemaining());
        dto.setCarriedOverLeave(employee.getCarriedOverLeave());
        return dto;
    }

    public static class CreateEmployeeRequest {
        private String name;
        private Long managerId;
        private LocalDate joinDate;

        public CreateEmployeeRequest() {
        }

        public String getName() {
            return name;
        }

        public void setName(String name) {
            this.name = name;
        }

        public Long getManagerId() {
            return managerId;
        }

        public void setManagerId(Long managerId) {
            this.managerId = managerId;
        }

        public LocalDate getJoinDate() {
            return joinDate;
        }

        public void setJoinDate(LocalDate joinDate) {
            this.joinDate = joinDate;
        }
    }
}
