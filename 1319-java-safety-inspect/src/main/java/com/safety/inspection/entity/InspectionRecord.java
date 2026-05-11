package com.safety.inspection.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.time.LocalDateTime;

@Data
@EqualsAndHashCode(callSuper = true)
@TableName("inspection_record")
public class InspectionRecord extends BaseEntity {

    private Long taskId;

    private Long checkItemId;

    private String checkResult;

    private String description;

    private String photos;

    private Long createdBy;

    private LocalDateTime createdAt;
}
