package com.example.proxy.dto;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class BackendStats {
    private String backendUrl;
    private String status;
    private int weight;
    private long totalRequests;
    private long successCount;
    private long errorCount;
    private double avgResponseTime;
}
