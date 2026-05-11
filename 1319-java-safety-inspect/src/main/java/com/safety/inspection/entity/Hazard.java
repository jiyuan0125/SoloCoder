package com.safety.inspection.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;
import lombok.EqualsAndHashCode;

import java.time.LocalDateTime;

@Data
@EqualsAndHashCode(callSuper = true)
@TableName("hazard")
public class Hazard extends BaseEntity {

    private String hazardNo;

    private Long inspectionRecordId;

    private Long taskId;

    private Long areaId;

    private String hazardLevel;

    private String originalLevel;

    private Integer upgradeCount;

    private String hazardType;

    private String location;

    private String description;

    private String photos;

    private String status;

    private LocalDateTime rectificationDeadline;

    private String rectificationPlan;

    private Long responsiblePersonId;

    private String rectificationResult;

    private String rectificationPhotos;

    private LocalDateTime rectificationTime;

    private Long recheckerId;

    private String recheckResult;

    private LocalDateTime recheckTime;

    private String recheckRemark;

    private Integer isRepeat;

    private Long relatedHazardId;

    private Long createdBy;

    private LocalDateTime closedAt;
}
