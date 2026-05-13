package com.example.metrics.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.Map;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class DataPoint {
    private String metricName;
    private MetricType type;
    private Map<String, String> labels;
    private double value;
    private Instant timestamp;
}
