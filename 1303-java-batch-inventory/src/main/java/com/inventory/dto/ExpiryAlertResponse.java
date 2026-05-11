package com.inventory.dto;

import lombok.Data;
import java.time.LocalDate;
import java.time.temporal.ChronoUnit;

@Data
public class ExpiryAlertResponse {
    private Long batchId;
    private String batchNumber;
    private Long productId;
    private String productName;
    private String productSku;
    private String category;
    private Integer currentQuantity;
    private LocalDate productionDate;
    private Integer shelfLifeDays;
    private LocalDate expiryDate;
    private Long daysUntilExpiry;
}
