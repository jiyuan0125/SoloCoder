package com.example.groupbuy.dto;

import lombok.Data;

import javax.validation.constraints.NotNull;

@Data
public class CreateGroupBuyDTO {
    
    @NotNull(message = "活动ID不能为空")
    private Long activityId;
    
    @NotNull(message = "用户ID不能为空")
    private Long userId;
    
    @NotNull(message = "目标成团人数档位不能为空")
    private Integer targetPeopleCount;
    
    @NotNull(message = "拼团有效期（小时）不能为空")
    private Integer validityHours;
}
