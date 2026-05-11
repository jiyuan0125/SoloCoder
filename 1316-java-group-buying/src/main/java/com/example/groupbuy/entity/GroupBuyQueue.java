package com.example.groupbuy.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("group_buy_queue")
public class GroupBuyQueue {
    
    @TableId(type = IdType.AUTO)
    private Long id;
    
    private Long groupOrderId;
    
    private Long activityId;
    
    private Long userId;
    
    private Integer queueStatus;
    
    private LocalDateTime queueTime;
    
    private LocalDateTime processTime;
    
    private LocalDateTime createTime;
}
