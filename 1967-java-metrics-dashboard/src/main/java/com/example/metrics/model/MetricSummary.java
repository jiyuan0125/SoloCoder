package com.example.metrics.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class MetricSummary {
    private String name;
    private String description;
    private Double latestValue;
    private Long updatedAt;
}
