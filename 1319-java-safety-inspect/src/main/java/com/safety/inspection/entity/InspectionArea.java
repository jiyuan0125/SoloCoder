package com.safety.inspection.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.EqualsAndHashCode;

@Data
@EqualsAndHashCode(callSuper = true)
@TableName("inspection_area")
public class InspectionArea extends BaseEntity {

    private String areaName;

    private String areaType;

    private String location;

    private String description;

    private Long managerId;

    private Long departmentId;

    private Integer status;
}
