package com.example.groupbuy.dto;

import lombok.Data;

import javax.validation.constraints.*;
import java.math.BigDecimal;

@Data
public class PriceTierDTO {
    
    private Long id;
    
    @NotNull(message = "最低成团人数不能为空")
    @Min(value = 2, message = "最低成团人数不能小于2")
    private Integer minPeopleCount;
    
    @NotNull(message = "目标人数档位不能为空")
    @Min(value = 2, message = "目标人数档位不能小于2")
    private Integer targetPeopleCount;
    
    @NotNull(message = "团购价不能为空")
    @DecimalMin(value = "0.01", message = "团购价必须大于0")
    private BigDecimal groupPrice;
    
    private BigDecimal discountRate;
    
    private Integer sortOrder;
}
