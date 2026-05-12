package com.logaggregator.dto;

import com.logaggregator.model.LogLevel;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.Map;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AggregationStats {

    private Map<String, Long> serviceDistribution;
    private Map<LogLevel, Long> levelDistribution;
    private Map<String, Long> errorTrendLastHour;
}
