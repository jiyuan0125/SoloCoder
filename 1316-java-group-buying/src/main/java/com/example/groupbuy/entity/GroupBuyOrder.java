package com.example.groupbuy.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@TableName("group_buy_order")
public class GroupBuyOrder {
    
    @TableId(type = IdType.AUTO)
    private Long id;
    
    private String orderNo;
    
    private Long activityId;
    
    private Long productId;
    
    private Long leaderUserId;
    
    private Integer targetPeopleCount;
    
    private Integer actualPeopleCount;
    
    private BigDecimal groupPrice;
    
    private Integer validityHours;
    
    private LocalDateTime startTime;
    
    private LocalDateTime endTime;
    
    private Integer status;
    
    private Integer queueOrder;
    
    private Integer isQueued;
    
    private LocalDateTime createTime;
    
    private LocalDateTime updateTime;
}
