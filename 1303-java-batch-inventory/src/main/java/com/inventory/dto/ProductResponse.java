package com.inventory.dto;

import lombok.Data;
import java.time.LocalDateTime;

@Data
public class ProductResponse {
    private Long id;
    private String name;
    private String sku;
    private String category;
    private LocalDateTime createdAt;
    private LocalDateTime updatedAt;
}
