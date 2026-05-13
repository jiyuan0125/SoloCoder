package com.example.metrics.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

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
}
