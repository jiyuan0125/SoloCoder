package com.safety.inspection.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.time.LocalDateTime;

@Data
@EqualsAndHashCode(callSuper = true)
@TableName("monthly_report")
public class MonthlyReport extends BaseEntity {

    private Integer reportYear;

    private Integer reportMonth;

    private Integer totalHazards;

    private Integer rectifiedCount;

    private Integer rectifyingCount;

    private Integer overdueCount;

    private Integer generalCount;

    private Integer largerCount;

    private Integer majorCount;

    private LocalDateTime generatedAt;
}
