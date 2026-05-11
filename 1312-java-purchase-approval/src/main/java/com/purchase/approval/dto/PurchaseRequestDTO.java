package com.purchase.approval.dto;

import lombok.Data;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Positive;
import java.math.BigDecimal;
import java.time.LocalDate;

@Data
public class PurchaseRequestDTO {
    @NotBlank(message = "采购物品名称不能为空")
    private String itemName;
    
    private String specification;
    
    @NotNull(message = "数量不能为空")
    @Positive(message = "数量必须大于0")
    private Integer quantity;
    
    @NotNull(message = "预估单价不能为空")
    @Positive(message = "预估单价必须大于0")
    private BigDecimal estimatedUnitPrice;
    
    private LocalDate expectedDeliveryDate;
    
    private Long recommendedSupplierId;
    
    @NotNull(message = "提交人ID不能为空")
    private Long submitterId;
    
    @NotBlank(message = "提交人姓名不能为空")
    private String submitterName;
}
