package com.example.proxy.dto;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.Map;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AccessLogStats {
    private long totalRequests;
    private long successCount;
    private long errorCount;
    private long total4xxCount;
    private long total5xxCount;
    private double avgResponseTime;
    private double p95ResponseTime;
    private Map<String, Long> requestsByRoute;
    private Map<String, Long> requestsByBackend;
}
