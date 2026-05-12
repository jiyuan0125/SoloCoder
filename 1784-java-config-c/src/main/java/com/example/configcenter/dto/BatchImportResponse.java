package com.example.configcenter.dto;

import lombok.Builder;
import lombok.Data;

import java.util.List;

@Data
@Builder
public class BatchImportResponse {
    private Integer total;
    private Integer successCount;
    private Integer failureCount;
    private List<FailureItem> failures;
    
    @Data
    @Builder
    public static class FailureItem {
        private String key;
        private String reason;
    }
}
