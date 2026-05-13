package com.example.metrics.model;

import lombok.Getter;

@Getter
public enum TimeWindow {
    ONE_MINUTE("1m", 60 * 1000, 24 * 60 * 60 * 1000),
    FIVE_MINUTES("5m", 5 * 60 * 1000, 7 * 24 * 60 * 60 * 1000),
    ONE_HOUR("1h", 60 * 60 * 1000, 30L * 24 * 60 * 60 * 1000);

    private final String name;
    private final long windowSize;
    private final long retention;

    TimeWindow(String name, long windowSize, long retention) {
        this.name = name;
        this.windowSize = windowSize;
        this.retention = retention;
    }

    public long alignToWindowStart(long timestamp) {
        return (timestamp / windowSize) * windowSize;
    }

    public long alignToWindowEnd(long timestamp) {
        return alignToWindowStart(timestamp) + windowSize - 1;
    }
}
