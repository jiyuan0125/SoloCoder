package com.inventory.dto;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;
import java.time.LocalDate;

@Data
public class InboundRequest {
    @NotNull(message = "商品ID不能为空")
    private Long productId;

    @NotBlank(message = "批次号不能为空")
    private String batchNumber;

    @NotNull(message = "入库日期不能为空")
    private LocalDate inboundDate;

    @NotNull(message = "生产日期不能为空")
    private LocalDate productionDate;

    @NotNull(message = "保质期天数不能为空")
    @Min(value = 1, message = "保质期天数必须大于0")
    private Integer shelfLifeDays;

    @NotNull(message = "入库数量不能为空")
    @Min(value = 1, message = "入库数量必须大于0")
    private Integer quantity;
}
