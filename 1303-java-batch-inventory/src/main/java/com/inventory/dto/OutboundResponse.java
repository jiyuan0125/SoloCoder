package com.inventory.dto;

import lombok.Data;
import java.time.LocalDateTime;
import java.util.List;

@Data
public class OutboundResponse {
    private Long productId;
    private String productName;
    private Integer totalQuantity;
    private LocalDateTime outboundTime;
    private List<OutboundDetail> details;

    @Data
    public static class OutboundDetail {
        private Long batchId;
        private String batchNumber;
        private Integer quantity;
    }
}
