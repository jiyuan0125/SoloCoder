package com.foodordering.entity;

import com.foodordering.enums.TasteOption;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class OrderItem {
    private Long dishId;
    private String dishName;
    private Double dishPrice;
    private Integer quantity;
    private TasteOption taste;
    private Double subtotal;
}
