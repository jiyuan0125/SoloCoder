package com.inventory.dto;

import lombok.Data;
import java.time.LocalDate;
import java.time.LocalDateTime;

@Data
public class BatchResponse {
    private Long id;
    private String batchNumber;
    private Long productId;
    private String productName;
    private LocalDate inboundDate;
    private LocalDate productionDate;
    private Integer shelfLifeDays;
    private LocalDate expiryDate;
    private Integer initialQuantity;
    private Integer currentQuantity;
    private LocalDateTime createdAt;
}
