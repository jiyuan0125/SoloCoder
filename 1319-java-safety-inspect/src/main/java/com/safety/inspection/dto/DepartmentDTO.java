package com.safety.inspection.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class DepartmentDTO {

    private Long id;

    @NotBlank(message = "部门名称不能为空")
    private String deptName;

    private Long parentId;

    private Long leaderId;

    private Integer sort;

    private Integer status;
}
