package com.safety.inspection.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.EqualsAndHashCode;

@Data
@EqualsAndHashCode(callSuper = true)
@TableName("inspection_check_item")
public class InspectionCheckItem extends BaseEntity {

    private Long areaId;

    private String itemName;

    private String itemDescription;

    private String standard;

    private String riskLevel;

    private Integer sort;

    private Integer status;
}
