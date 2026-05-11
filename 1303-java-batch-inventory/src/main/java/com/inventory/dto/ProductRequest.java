package com.inventory.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class ProductRequest {
    @NotBlank(message = "商品名称不能为空")
    private String name;

    @NotBlank(message = "SKU编码不能为空")
    private String sku;

    @NotBlank(message = "商品类别不能为空")
    private String category;
}
