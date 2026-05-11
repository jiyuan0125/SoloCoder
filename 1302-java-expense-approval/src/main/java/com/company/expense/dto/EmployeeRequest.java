package com.company.expense.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.Size;

@Data
public class EmployeeRequest {

    @NotBlank(message = "员工编号不能为空")
    @Size(max = 50, message = "员工编号长度不能超过50个字符")
    private String employeeId;

    @NotBlank(message = "员工姓名不能为空")
    @Size(max = 100, message = "员工姓名长度不能超过100个字符")
    private String name;

    @NotBlank(message = "部门不能为空")
    @Size(max = 100, message = "部门长度不能超过100个字符")
    private String department;

    private String supervisorEmployeeId;

    private Boolean isManager = false;

    private Boolean isGeneralManager = false;
}
