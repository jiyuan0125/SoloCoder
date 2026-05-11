package com.foodordering.dto.request;

import com.foodordering.enums.Category;
import com.foodordering.enums.TasteOption;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.Positive;
import lombok.Data;

import java.util.List;

@Data
public class UpdateDishRequest {
    private String name;
    
    @Positive(message = "价格必须大于0")
    private Double price;
    
    private Category category;
    
    private List<TasteOption> availableTastes;
    
    @Min(value = 0, message = "库存不能小于0")
    private Integer stock;
    
    private Boolean isActive;
}
