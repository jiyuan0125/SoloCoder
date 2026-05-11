package com.purchase.approval.dto;

import lombok.Data;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Positive;
import java.math.BigDecimal;

@Data
public class QuoteDTO {
    @NotNull(message = "供应商ID不能为空")
    private Long supplierId;
    
    @NotNull(message = "单价不能为空")
    @Positive(message = "单价必须大于0")
    private BigDecimal unitPrice;
    
    private java.time.LocalDateTime deliveryDate;
    
    private String paymentTerms;
    
    private String remarks;
    
    private String submittedBy;
}
