package com.safety.inspection.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("hazard_upgrade_record")
public class HazardUpgradeRecord {

    private Long id;

    private Long hazardId;

    private String oldLevel;

    private String newLevel;

    private String reason;

    private LocalDateTime newDeadline;

    private LocalDateTime createdAt;
}
