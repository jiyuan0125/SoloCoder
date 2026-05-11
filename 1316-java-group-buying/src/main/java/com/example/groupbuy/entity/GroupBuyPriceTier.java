package com.example.groupbuy.entity;

import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.annotation.TableId;
import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@TableName("group_buy_price_tier")
public class GroupBuyPriceTier {
    
    @TableId(type = IdType.AUTO)
    private Long id;
    
    private Long activityId;
    
    private Integer minPeopleCount;
    
    private Integer targetPeopleCount;
    
    private BigDecimal groupPrice;
    
    private BigDecimal discountRate;
    
    private Integer sortOrder;
    
    private LocalDateTime createTime;
}
