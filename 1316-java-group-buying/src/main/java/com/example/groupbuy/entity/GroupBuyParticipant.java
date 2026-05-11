package com.example.groupbuy.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@TableName("group_buy_participant")
public class GroupBuyParticipant {
    
    @TableId(type = IdType.AUTO)
    private Long id;
    
    private Long groupOrderId;
    
    private Long userId;
    
    private Integer isLeader;
    
    private LocalDateTime joinTime;
    
    private LocalDateTime payTime;
    
    private Integer payStatus;
    
    private BigDecimal payAmount;
    
    private BigDecimal refundAmount;
    
    private LocalDateTime refundTime;
    
    private Integer productDelivered;
    
    private LocalDateTime createTime;
    
    private LocalDateTime updateTime;
}
