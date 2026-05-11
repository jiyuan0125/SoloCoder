package com.foodordering.dto.request;

import com.foodordering.enums.Category;
import com.foodordering.enums.TasteOption;
import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import jakarta.validation.constraints.Positive;
import lombok.Data;

import java.util.List;

@Data
public class CreateDishRequest {
    @NotBlank(message = "菜品名称不能为空")
    private String name;
    
    @NotNull(message = "价格不能为空")
    @Positive(message = "价格必须大于0")
    private Double price;
    
    @NotNull(message = "分类不能为空")
    private Category category;
    
    private List<TasteOption> availableTastes;
    
    @NotNull(message = "库存不能为空")
    @Min(value = 0, message = "库存不能小于0")
    private Integer stock;
}
