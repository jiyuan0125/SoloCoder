package com.purchase.approval.dto;

import lombok.Data;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Positive;

@Data
public class InspectionDTO {
    @NotNull(message = "验收数量不能为空")
    @Positive(message = "验收数量必须大于0")
    private Integer inspectedQuantity;
    
    @NotNull(message = "合格数量不能为空")
    private Integer passedQuantity;
    
    private String inspectionRemarks;
    
    private String inspectedBy;
}
