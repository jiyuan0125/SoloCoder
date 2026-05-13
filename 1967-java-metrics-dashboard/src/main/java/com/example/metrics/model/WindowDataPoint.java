package com.example.metrics.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class WindowDataPoint {
    private long windowStart;
    private double sum;
    private double max;
    private double min;
    private long count;
    private double latestValue;

    public void addValue(double value) {
        sum += value;
        if (count == 0) {
            max = value;
            min = value;
        } else {
            max = Math.max(max, value);
            min = Math.min(min, value);
        }
        count++;
        latestValue = value;
    }

    public double getAvg() {
        return count > 0 ? sum / count : 0;
    }
}
