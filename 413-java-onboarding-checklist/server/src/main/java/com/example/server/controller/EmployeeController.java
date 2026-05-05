package com.example.server.controller;

import com.example.common.dto.ChecklistItemDTO;
import com.example.common.dto.EmployeeDTO;
import com.example.common.enums.ErrorCode;
import com.example.common.request.CreateChecklistItemRequest;
import com.example.common.request.CreateEmployeeRequest;
import com.example.common.request.UpdateChecklistItemRequest;
import com.example.common.response.ApiResponse;
import com.example.server.service.EmployeeService;
import jakarta.validation.Valid;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/api/employees")
public class EmployeeController {

    @Autowired
    private EmployeeService employeeService;

    @PostMapping
    public ApiResponse<EmployeeDTO> createEmployee(@Valid @RequestBody CreateEmployeeRequest request) {
        EmployeeDTO employee = employeeService.createEmployee(request);
        return ApiResponse.success(employee);
    }

    @GetMapping
    public ApiResponse<List<EmployeeDTO>> getAllEmployees() {
        List<EmployeeDTO> employees = employeeService.getAllEmployees();
        return ApiResponse.success(employees);
    }

    @GetMapping("/{id}")
    public ApiResponse<EmployeeDTO> getEmployeeById(@PathVariable String id) {
        Optional<EmployeeDTO> employee = employeeService.getEmployeeById(id);
        if (employee.isPresent()) {
            return ApiResponse.success(employee.get());
        }
        return ApiResponse.error(ErrorCode.EMPLOYEE_NOT_FOUND);
    }

    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteEmployee(@PathVariable String id) {
        boolean deleted = employeeService.deleteEmployee(id);
        if (deleted) {
            return ApiResponse.success();
        }
        return ApiResponse.error(ErrorCode.EMPLOYEE_NOT_FOUND);
    }

    @PostMapping("/{employeeId}/items")
    public ApiResponse<ChecklistItemDTO> addChecklistItem(
            @PathVariable String employeeId,
            @Valid @RequestBody CreateChecklistItemRequest request) {
        ChecklistItemDTO item = new ChecklistItemDTO();
        item.setName(request.getName());
        item.setDescription(request.getDescription());
        item.setResponsiblePerson(request.getResponsiblePerson());
        item.setDepartment(request.getDepartment());
        item.setDueDate(request.getDueDate());
        item.setRequired(request.isRequired());
        item.setOnboardingDayRequired(request.isOnboardingDayRequired());
        item.setPreOnboarding(request.isPreOnboarding());

        Optional<ChecklistItemDTO> addedItem = employeeService.addChecklistItem(employeeId, item);
        if (addedItem.isPresent()) {
            return ApiResponse.success(addedItem.get());
        }
        return ApiResponse.error(ErrorCode.EMPLOYEE_NOT_FOUND);
    }

    @PutMapping("/{employeeId}/items/{itemId}")
    public ApiResponse<ChecklistItemDTO> updateChecklistItem(
            @PathVariable String employeeId,
            @PathVariable String itemId,
            @RequestBody UpdateChecklistItemRequest request) {
        Optional<ChecklistItemDTO> updatedItem = employeeService.updateChecklistItem(
                employeeId, itemId,
                request.getName(),
                request.getDescription(),
                request.getResponsiblePerson(),
                request.getDepartment(),
                request.getDueDate(),
                request.getIsCompleted()
        );
        if (updatedItem.isPresent()) {
            return ApiResponse.success(updatedItem.get());
        }
        return ApiResponse.error(ErrorCode.CHECKLIST_ITEM_NOT_FOUND);
    }
}
