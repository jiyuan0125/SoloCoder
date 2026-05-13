package com.example.metrics.storage;

import com.example.metrics.model.MetricPoint;
import com.example.metrics.model.MetricRegistration;
import com.example.metrics.model.TimeWindow;
import com.example.metrics.model.WindowDataPoint;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.TreeMap;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Component
public class MetricStorage {

    private final Map<String, MetricRegistration> registrations = new ConcurrentHashMap<>();
    private final Map<String, Long> latestUpdateTimes = new ConcurrentHashMap<>();
    private final Map<String, Double> latestValues = new ConcurrentHashMap<>();
    private final Map<String, Map<String, Map<Long, WindowDataPoint>>> data = new ConcurrentHashMap<>();

    public synchronized void registerMetric(MetricRegistration registration) {
        registrations.put(registration.getName(), registration);
    }

    public boolean isRegistered(String name) {
        return registrations.containsKey(name);
    }

    public MetricRegistration getRegistration(String name) {
        return registrations.get(name);
    }

    public List<MetricRegistration> getAllRegistrations() {
        return new ArrayList<>(registrations.values());
    }

    public synchronized void storeMetric(MetricPoint point) {
        long timestamp = point.getTimestamp();
        String name = point.getMetricName();
        String labelKey = getLabelKey(point.getLabels());

        for (TimeWindow window : TimeWindow.values()) {
            String windowName = window.getName();
            long windowStart = window.alignToWindowStart(timestamp);

            Map<String, Map<Long, WindowDataPoint>> windowData = data
                    .computeIfAbsent(name, k -> new ConcurrentHashMap<>());
            Map<Long, WindowDataPoint> labelData = windowData
                    .computeIfAbsent(labelKey, k -> new ConcurrentHashMap<>());
            WindowDataPoint dataPoint = labelData
                    .computeIfAbsent(windowStart, k -> WindowDataPoint.builder()
                            .windowStart(windowStart)
                            .sum(0)
                            .max(Double.MIN_VALUE)
                            .min(Double.MAX_VALUE)
                            .count(0)
                            .latestValue(0)
                            .build());
            dataPoint.addValue(point.getValue());

            cleanupOldData(name, labelKey, window, timestamp);
        }

        latestValues.put(name, point.getValue());
        latestUpdateTimes.put(name, timestamp);
    }

    public List<WindowDataPoint> queryWindowData(String name, String labelKey, TimeWindow window, long startTime, long endTime) {
        Map<String, Map<Long, WindowDataPoint>> windowData = data.get(name);
        if (windowData == null) {
            return Collections.emptyList();
        }
        Map<Long, WindowDataPoint> labelData = windowData.get(labelKey);
        if (labelData == null) {
            return Collections.emptyList();
        }
        long windowStart = window.alignToWindowStart(startTime);
        long windowEnd = window.alignToWindowEnd(endTime);
        return labelData.entrySet().stream()
                .filter(e -> e.getKey() >= windowStart && e.getKey() <= windowEnd)
                .map(Map.Entry::getValue)
                .sorted((a, b) -> Long.compare(a.getWindowStart(), b.getWindowStart()))
                .collect(Collectors.toList());
    }

    public WindowDataPoint getWindowData(String name, String labelKey, TimeWindow window, long timestamp) {
        Map<String, Map<Long, WindowDataPoint>> windowData = data.get(name);
        if (windowData == null) {
            return null;
        }
        Map<Long, WindowDataPoint> labelData = windowData.get(labelKey);
        if (labelData == null) {
            return null;
        }
        long windowStart = window.alignToWindowStart(timestamp);
        return labelData.get(windowStart);
    }

    public Double getLatestValue(String name) {
        return latestValues.get(name);
    }

    public Long getLatestUpdateTime(String name) {
        return latestUpdateTimes.get(name);
    }

    private void cleanupOldData(String name, String labelKey, TimeWindow window, long currentTime) {
        long cutoffTime = currentTime - window.getRetention();
        Map<String, Map<Long, WindowDataPoint>> windowData = data.get(name);
        if (windowData == null) {
            return;
        }
        Map<Long, WindowDataPoint> labelData = windowData.get(labelKey);
        if (labelData == null) {
            return;
        }
        labelData.keySet().removeIf(ws -> ws < cutoffTime);
    }

    private String getLabelKey(Map<String, String> labels) {
        if (labels == null || labels.isEmpty()) {
            return "";
        }
        TreeMap<String, String> sorted = new TreeMap<>(labels);
        return sorted.entrySet().stream()
                .map(e -> e.getKey() + "=" + e.getValue())
                .collect(Collectors.joining(","));
    }

    public List<String> getAllLabelKeys(String name) {
        Map<String, Map<Long, WindowDataPoint>> windowData = data.get(name);
        if (windowData == null) {
            return Collections.emptyList();
        }
        return new ArrayList<>(windowData.keySet());
    }

    public TimeWindow getBestGranularityForRange(long range) {
        if (range <= TimeWindow.ONE_MINUTE.getRetention()) {
            return TimeWindow.ONE_MINUTE;
        } else if (range <= TimeWindow.FIVE_MINUTES.getRetention()) {
            return TimeWindow.FIVE_MINUTES;
        } else {
            return TimeWindow.ONE_HOUR;
        }
    }
}
