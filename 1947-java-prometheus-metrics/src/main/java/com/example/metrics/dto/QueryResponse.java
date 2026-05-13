package com.example.metrics.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.List;
import java.util.Map;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class QueryResponse {
    private String metricName;
    private Map<String, String> labels;
    private Double p50;
    private Double p95;
    private Double p99;
    private Double average;
    private Long count;
    private List<TimeWindow> windows;

    @Data
    @AllArgsConstructor
    @NoArgsConstructor
    public static class TimeWindow {
        private Instant windowStart;
        private Instant windowEnd;
        private Double p50;
        private Double p95;
        private Double p99;
        private Double average;
        private Long count;
    }
}
