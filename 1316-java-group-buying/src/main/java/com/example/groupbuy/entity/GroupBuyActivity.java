package com.example.groupbuy.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@TableName("group_buy_activity")
public class GroupBuyActivity {
    
    @TableId(type = IdType.AUTO)
    private Long id;
    
    private Long productId;
    
    private String productName;
    
    private BigDecimal originalPrice;
    
    private Integer productType;
    
    private Integer maxGroupCount;
    
    private Integer usedGroupCount;
    
    private LocalDateTime startTime;
    
    private LocalDateTime endTime;
    
    private Integer status;
    
    private LocalDateTime createTime;
    
    private LocalDateTime updateTime;
}
