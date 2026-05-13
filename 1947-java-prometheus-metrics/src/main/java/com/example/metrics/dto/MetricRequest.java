package com.example.metrics.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

import java.util.Map;

@Data
public class MetricRequest {
    @NotBlank(message = "metricName is required")
    private String metricName;
    private Map<String, String> labels;
    @NotNull(message = "value is required")
    private Double value;
}
