package com.inventory.dto;

import lombok.Data;
import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
public class InventoryResponse {
    private Long id;
    private Long batchId;
    private String batchNumber;
    private Long productId;
    private String productName;
    private String productSku;
    private String category;
    private Integer currentQuantity;
    private LocalDate inboundDate;
    private LocalDate productionDate;
    private LocalDate expiryDate;
    private LocalDateTime updatedAt;
}
