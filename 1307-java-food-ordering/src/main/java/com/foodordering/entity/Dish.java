package com.foodordering.entity;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.foodordering.enums.Category;
import com.foodordering.enums.TasteOption;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class Dish {
    private Long id;
    private String name;
    private Double price;
    private Category category;
    private List<TasteOption> availableTastes;
    private Integer stock;
    private Boolean isActive;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
    
    public Boolean getSoldOut() {
        return stock == null || stock <= 0;
    }
}
