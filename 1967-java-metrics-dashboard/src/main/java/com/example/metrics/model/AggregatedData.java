package com.example.metrics.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class AggregatedData {
    private Long windowStart;
    private Long windowEnd;
    private String granularity;
    private Double avg;
    private Double max;
    private Double min;
    private Long count;
    private Double sum;
    private Double YoYChange;
    private Double MoMChange;
    private Boolean anomaly;
}
