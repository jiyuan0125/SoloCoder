package com.example.groupbuy.dto;

import lombok.Data;

import javax.validation.Valid;
import javax.validation.constraints.*;
import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;

@Data
public class GroupBuyActivityDTO {
    
    private Long id;
    
    @NotNull(message = "商品ID不能为空")
    private Long productId;
    
    @NotBlank(message = "商品名称不能为空")
    private String productName;
    
    @NotNull(message = "原价不能为空")
    @DecimalMin(value = "0.01", message = "原价必须大于0")
    private BigDecimal originalPrice;
    
    @NotNull(message = "商品类型不能为空")
    @Min(value = 1, message = "商品类型错误")
    @Max(value = 2, message = "商品类型错误")
    private Integer productType;
    
    @NotNull(message = "最大团数不能为空")
    @Min(value = 1, message = "最大团数必须大于0")
    private Integer maxGroupCount;
    
    @NotNull(message = "活动开始时间不能为空")
    @Future(message = "活动开始时间必须是未来时间")
    private LocalDateTime startTime;
    
    @NotNull(message = "活动结束时间不能为空")
    private LocalDateTime endTime;
    
    @Valid
    @NotNull(message = "阶梯价格不能为空")
    @Size(min = 1, message = "至少需要一个阶梯价格")
    private List<PriceTierDTO> priceTiers;
}
