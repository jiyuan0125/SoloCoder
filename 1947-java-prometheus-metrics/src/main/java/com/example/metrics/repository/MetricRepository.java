package com.example.metrics.repository;

import com.example.metrics.config.MetricsProperties;
import com.example.metrics.exception.LabelLimitExceededException;
import com.example.metrics.model.DataPoint;
import com.example.metrics.model.MetricType;
import org.springframework.stereotype.Repository;

import java.time.Instant;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.stream.Collectors;

@Repository
public class MetricRepository {

    private final MetricsProperties properties;
    private final ConcurrentMap<String, ConcurrentMap<String, List<DataPoint>>> rawData = new ConcurrentHashMap<>();
    private final ConcurrentMap<String, Set<String>> labelCombinations = new ConcurrentHashMap<>();
    private final ConcurrentMap<String, ConcurrentMap<String, Double>> currentValues = new ConcurrentHashMap<>();
    private final ConcurrentMap<String, ConcurrentMap<String, List<Double>>> histogramBuckets = new ConcurrentHashMap<>();

    public MetricRepository(MetricsProperties properties) {
        this.properties = properties;
    }

    public synchronized void addDataPoint(DataPoint dataPoint) {
        String metricName = dataPoint.getMetricName();
        String labelKey = getLabelKey(dataPoint.getLabels());
        String metricKey = metricName + ":" + dataPoint.getType().name();

        labelCombinations.computeIfAbsent(metricKey, k -> Collections.newSetFromMap(new ConcurrentHashMap<>()));
        Set<String> combinations = labelCombinations.get(metricKey);

        if (!combinations.contains(labelKey) && combinations.size() >= properties.getLabels().getMaxCombinations()) {
            throw new LabelLimitExceededException(
                String.format("Metric '%s' has exceeded maximum label combinations (%d)",
                    metricName, properties.getLabels().getMaxCombinations())
            );
        }

        combinations.add(labelKey);

        rawData.computeIfAbsent(metricKey, k -> new ConcurrentHashMap<>());
        rawData.get(metricKey).computeIfAbsent(labelKey, k -> new ArrayList<>());

        cleanOldData(metricKey, labelKey);

        List<DataPoint> points = rawData.get(metricKey).get(labelKey);
        synchronized (points) {
            if (dataPoint.getType() == MetricType.COUNTER) {
                if (points.isEmpty()) {
                    points.add(new DataPoint(metricName, MetricType.COUNTER, dataPoint.getLabels(), 
                        Math.max(0, dataPoint.getValue()), dataPoint.getTimestamp()));
                } else {
                    DataPoint last = points.get(points.size() - 1);
                    points.add(new DataPoint(metricName, MetricType.COUNTER, dataPoint.getLabels(),
                        Math.max(last.getValue(), dataPoint.getValue()), dataPoint.getTimestamp()));
                }
            } else {
                points.add(dataPoint);
            }
        }

        updateCurrentValues(dataPoint, labelKey);
    }

    private void updateCurrentValues(DataPoint dataPoint, String labelKey) {
        String metricKey = dataPoint.getMetricName() + ":" + dataPoint.getType().name();
        currentValues.computeIfAbsent(metricKey, k -> new ConcurrentHashMap<>());
        
        if (dataPoint.getType() == MetricType.COUNTER) {
            Double current = currentValues.get(metricKey).getOrDefault(labelKey, 0.0);
            currentValues.get(metricKey).put(labelKey, Math.max(current, dataPoint.getValue()));
        } else if (dataPoint.getType() == MetricType.GAUGE) {
            currentValues.get(metricKey).put(labelKey, dataPoint.getValue());
        } else if (dataPoint.getType() == MetricType.HISTOGRAM) {
            updateHistogramBucket(dataPoint, labelKey);
        }
    }

    private void updateHistogramBucket(DataPoint dataPoint, String labelKey) {
        String metricKey = dataPoint.getMetricName() + ":" + MetricType.HISTOGRAM.name();
        histogramBuckets.computeIfAbsent(metricKey, k -> new ConcurrentHashMap<>());
        histogramBuckets.get(metricKey).computeIfAbsent(labelKey, k -> {
            List<Double> buckets = new ArrayList<>();
            for (int i = 0; i < properties.getBuckets().getHistogram().size() + 2; i++) {
                buckets.add(0.0);
            }
            return buckets;
        });

        List<Double> buckets = histogramBuckets.get(metricKey).get(labelKey);
        double value = dataPoint.getValue();
        List<Double> bucketUpperBounds = properties.getBuckets().getHistogram();

        synchronized (buckets) {
            for (int i = 0; i < bucketUpperBounds.size(); i++) {
                if (value <= bucketUpperBounds.get(i)) {
                    buckets.set(i, buckets.get(i) + 1);
                    break;
                }
            }
            if (value > bucketUpperBounds.get(bucketUpperBounds.size() - 1)) {
                buckets.set(bucketUpperBounds.size(), buckets.get(bucketUpperBounds.size()) + 1);
            }
            buckets.set(bucketUpperBounds.size() + 1, buckets.get(bucketUpperBounds.size() + 1) + value);
        }
    }

    public List<DataPoint> queryRawData(String metricName, MetricType type, Map<String, String> labels, 
                                        Instant startTime, Instant endTime) {
        String metricKey = metricName + ":" + type.name();
        if (!rawData.containsKey(metricKey)) {
            return Collections.emptyList();
        }

        String labelKey = getLabelKey(labels);
        ConcurrentMap<String, List<DataPoint>> metricData = rawData.get(metricKey);
        
        List<DataPoint> result = new ArrayList<>();
        
        for (Map.Entry<String, List<DataPoint>> entry : metricData.entrySet()) {
            if (labelKey == null || entry.getKey().equals(labelKey)) {
                synchronized (entry.getValue()) {
                    for (DataPoint dp : entry.getValue()) {
                        if ((startTime == null || !dp.getTimestamp().isBefore(startTime)) &&
                            (endTime == null || !dp.getTimestamp().isAfter(endTime))) {
                            result.add(dp);
                        }
                    }
                }
            }
        }
        
        return result;
    }

    public Map<String, Double> getCurrentValues(String metricName, MetricType type) {
        String metricKey = metricName + ":" + type.name();
        if (!currentValues.containsKey(metricKey)) {
            return Collections.emptyMap();
        }
        return new HashMap<>(currentValues.get(metricKey));
    }

    public Map<String, List<Double>> getHistogramBuckets(String metricName) {
        String metricKey = metricName + ":" + MetricType.HISTOGRAM.name();
        if (!histogramBuckets.containsKey(metricKey)) {
            return Collections.emptyMap();
        }
        Map<String, List<Double>> result = new HashMap<>();
        for (Map.Entry<String, List<Double>> entry : histogramBuckets.get(metricKey).entrySet()) {
            synchronized (entry.getValue()) {
                result.put(entry.getKey(), new ArrayList<>(entry.getValue()));
            }
        }
        return result;
    }

    public Set<String> getAllMetricKeys() {
        return currentValues.keySet();
    }

    public Map<String, String> parseLabelKey(String labelKey) {
        if (labelKey == null || labelKey.isEmpty()) {
            return Collections.emptyMap();
        }
        Map<String, String> labels = new HashMap<>();
        for (String part : labelKey.split(",")) {
            String[] kv = part.split("=", 2);
            if (kv.length == 2) {
                labels.put(kv[0], kv[1]);
            }
        }
        return labels;
    }

    private void cleanOldData(String metricKey, String labelKey) {
        if (!rawData.containsKey(metricKey) || !rawData.get(metricKey).containsKey(labelKey)) {
            return;
        }
        List<DataPoint> points = rawData.get(metricKey).get(labelKey);
        Instant cutoff = Instant.now().minusSeconds(properties.getRetention().getRawDataHours() * 3600L);
        
        synchronized (points) {
            int removeFrom = 0;
            for (int i = 0; i < points.size(); i++) {
                if (points.get(i).getTimestamp().isAfter(cutoff)) {
                    removeFrom = i;
                    break;
                }
            }
            if (removeFrom > 0) {
                points.subList(0, removeFrom).clear();
            }
        }
    }

    private String getLabelKey(Map<String, String> labels) {
        if (labels == null || labels.isEmpty()) {
            return "";
        }
        return labels.entrySet().stream()
            .sorted(Map.Entry.comparingByKey())
            .map(e -> e.getKey() + "=" + e.getValue())
            .collect(Collectors.joining(","));
    }
}
