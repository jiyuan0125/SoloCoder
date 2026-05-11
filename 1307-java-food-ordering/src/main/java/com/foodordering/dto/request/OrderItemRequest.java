package com.foodordering.dto.request;

import com.foodordering.enums.TasteOption;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class OrderItemRequest {
    @NotNull(message = "菜品ID不能为空")
    private Long dishId;
    
    @NotNull(message = "数量不能为空")
    @Min(value = 1, message = "数量必须大于0")
    private Integer quantity;
    
    private TasteOption taste;
}
