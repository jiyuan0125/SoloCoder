package com.company.expense.dto;

import com.company.expense.entity.Employee;
import lombok.Builder;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@Builder
public class EmployeeResponse {

    private Long id;
    private String employeeId;
    private String name;
    private String department;
    private String supervisorEmployeeId;
    private String supervisorName;
    private Boolean isManager;
    private Boolean isGeneralManager;
    private LocalDateTime createdAt;

    public static EmployeeResponse fromEntity(Employee employee) {
        EmployeeResponseBuilder builder = EmployeeResponse.builder()
                .id(employee.getId())
                .employeeId(employee.getEmployeeId())
                .name(employee.getName())
                .department(employee.getDepartment())
                .isManager(employee.getIsManager())
                .isGeneralManager(employee.getIsGeneralManager())
                .createdAt(employee.getCreatedAt());

        if (employee.getSupervisor() != null) {
            builder.supervisorEmployeeId(employee.getSupervisor().getEmployeeId())
                   .supervisorName(employee.getSupervisor().getName());
        }

        return builder.build();
    }
}
