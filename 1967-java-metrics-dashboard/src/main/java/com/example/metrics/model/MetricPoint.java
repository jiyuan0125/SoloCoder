package com.example.metrics.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import java.util.Map;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class MetricPoint {
    @NotBlank(message = "metric_name is required")
    private String metricName;

    @NotNull(message = "timestamp is required")
    private Long timestamp;

    @NotNull(message = "value is required")
    private Double value;

    private Map<String, String> labels;
}
