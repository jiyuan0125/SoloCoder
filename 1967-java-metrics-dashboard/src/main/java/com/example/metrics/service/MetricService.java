package com.example.metrics.service;

import com.example.metrics.model.AggregatedData;
import com.example.metrics.model.MetricPoint;
import com.example.metrics.model.MetricRegistration;
import com.example.metrics.model.MetricSummary;
import com.example.metrics.model.TimeWindow;
import com.example.metrics.model.WindowDataPoint;
import com.example.metrics.storage.MetricStorage;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@Service
public class MetricService {

    private final MetricStorage storage;

    @Value("${metrics.anomaly-threshold:50.0}")
    private double anomalyThreshold;

    private static final long ONE_WEEK_MS = Duration.ofDays(7).toMillis();

    public MetricService(MetricStorage storage) {
        this.storage = storage;
    }

    public void registerMetric(MetricRegistration registration) {
        storage.registerMetric(registration);
    }

    public void recordMetric(MetricPoint point) {
        if (!storage.isRegistered(point.getMetricName())) {
            storage.registerMetric(new MetricRegistration(point.getMetricName(), null));
        }
        storage.storeMetric(point);
    }

    public List<MetricSummary> listAllMetrics() {
        return storage.getAllRegistrations().stream()
                .map(reg -> MetricSummary.builder()
                        .name(reg.getName())
                        .description(reg.getDescription())
                        .latestValue(storage.getLatestValue(reg.getName()))
                        .updatedAt(storage.getLatestUpdateTime(reg.getName()))
                        .build())
                .collect(Collectors.toList());
    }

    public List<AggregatedData> queryMetrics(String name, Long startTime, Long endTime, String granularity) {
        if (!storage.isRegistered(name)) {
            return new ArrayList<>();
        }

        long now = System.currentTimeMillis();
        long end = endTime != null ? endTime : now;
        long start = startTime != null ? startTime : end - Duration.ofHours(1).toMillis();
        long range = end - start;

        TimeWindow window = granularity != null ? getTimeWindowByName(granularity) : storage.getBestGranularityForRange(range);

        List<AggregatedData> result = new ArrayList<>();

        for (String labelKey : storage.getAllLabelKeys(name)) {
            List<WindowDataPoint> points = storage.queryWindowData(name, labelKey, window, start, end);
            for (WindowDataPoint point : points) {
                result.add(buildAggregatedData(name, labelKey, point, window));
            }
        }

        return result;
    }

    public List<AggregatedData> getAnomalies(String name) {
        if (!storage.isRegistered(name)) {
            return new ArrayList<>();
        }

        List<AggregatedData> allData = queryMetrics(name, null, null, null);
        return allData.stream()
                .filter(d -> Boolean.TRUE.equals(d.getAnomaly()))
                .collect(Collectors.toList());
    }

    private AggregatedData buildAggregatedData(String name, String labelKey, WindowDataPoint current, TimeWindow window) {
        long windowEnd = current.getWindowStart() + window.getWindowSize() - 1;

        Double yoyChange = calculateYoYChange(name, labelKey, current, window);
        Double momChange = calculateMoMChange(name, labelKey, current, window);

        boolean anomaly = false;
        if (yoyChange != null && Math.abs(yoyChange) > anomalyThreshold) {
            anomaly = true;
        }
        if (momChange != null && Math.abs(momChange) > anomalyThreshold) {
            anomaly = true;
        }

        return AggregatedData.builder()
                .windowStart(current.getWindowStart())
                .windowEnd(windowEnd)
                .granularity(window.getName())
                .avg(current.getAvg())
                .max(current.getMax())
                .min(current.getMin())
                .count(current.getCount())
                .sum(current.getSum())
                .YoYChange(yoyChange)
                .MoMChange(momChange)
                .anomaly(anomaly)
                .build();
    }

    private Double calculateYoYChange(String name, String labelKey, WindowDataPoint current, TimeWindow window) {
        long compareTimestamp = current.getWindowStart() - ONE_WEEK_MS;
        WindowDataPoint previous = storage.getWindowData(name, labelKey, window, compareTimestamp);
        if (previous == null || previous.getAvg() == 0) {
            return null;
        }
        return calculatePercentageChange(current.getAvg(), previous.getAvg());
    }

    private Double calculateMoMChange(String name, String labelKey, WindowDataPoint current, TimeWindow window) {
        long compareTimestamp = current.getWindowStart() - window.getWindowSize();
        WindowDataPoint previous = storage.getWindowData(name, labelKey, window, compareTimestamp);
        if (previous == null || previous.getAvg() == 0) {
            return null;
        }
        return calculatePercentageChange(current.getAvg(), previous.getAvg());
    }

    private double calculatePercentageChange(double current, double previous) {
        return ((current - previous) / Math.abs(previous)) * 100.0;
    }

    private TimeWindow getTimeWindowByName(String name) {
        for (TimeWindow window : TimeWindow.values()) {
            if (window.getName().equals(name)) {
                return window;
            }
        }
        throw new IllegalArgumentException("Unknown granularity: " + name);
    }

    public boolean isRegistered(String name) {
        return storage.isRegistered(name);
    }

    public boolean exists(String name) {
        return storage.isRegistered(name) || storage.hasData(name);
    }
}
