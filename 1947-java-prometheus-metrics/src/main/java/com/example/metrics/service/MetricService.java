package com.example.metrics.service;

import com.example.metrics.config.MetricsProperties;
import com.example.metrics.dto.QueryResponse;
import com.example.metrics.model.DataPoint;
import com.example.metrics.model.MetricType;
import com.example.metrics.repository.MetricRepository;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.*;
import java.util.stream.Collectors;

@Service
public class MetricService {

    private final MetricRepository repository;
    private final MetricsProperties properties;

    public MetricService(MetricRepository repository, MetricsProperties properties) {
        this.repository = repository;
        this.properties = properties;
    }

    public void recordCounter(String metricName, Map<String, String> labels, double value) {
        if (value < 0) {
            throw new IllegalArgumentException("Counter value cannot be negative");
        }
        DataPoint dp = new DataPoint(metricName, MetricType.COUNTER, labels, value, Instant.now());
        repository.addDataPoint(dp);
    }

    public void recordGauge(String metricName, Map<String, String> labels, double value) {
        DataPoint dp = new DataPoint(metricName, MetricType.GAUGE, labels, value, Instant.now());
        repository.addDataPoint(dp);
    }

    public void recordHistogram(String metricName, Map<String, String> labels, double value) {
        DataPoint dp = new DataPoint(metricName, MetricType.HISTOGRAM, labels, value, Instant.now());
        repository.addDataPoint(dp);
    }

    public String exportPrometheusFormat() {
        StringBuilder sb = new StringBuilder();
        Set<String> metricKeys = repository.getAllMetricKeys();

        Map<String, List<String>> groupedMetrics = new HashMap<>();
        for (String key : metricKeys) {
            String[] parts = key.split(":", 2);
            if (parts.length == 2) {
                groupedMetrics.computeIfAbsent(parts[0], k -> new ArrayList<>()).add(parts[1]);
            }
        }

        for (Map.Entry<String, List<String>> entry : groupedMetrics.entrySet()) {
            String metricName = entry.getKey();
            List<String> types = entry.getValue();

            for (String typeStr : types) {
                MetricType type = MetricType.valueOf(typeStr);
                
                if (type == MetricType.COUNTER) {
                    sb.append("# HELP ").append(metricName).append("_total ").append(metricName).append(" counter\n");
                    sb.append("# TYPE ").append(metricName).append("_total counter\n");
                    Map<String, Double> values = repository.getCurrentValues(metricName, type);
                    for (Map.Entry<String, Double> valEntry : values.entrySet()) {
                        Map<String, String> labels = repository.parseLabelKey(valEntry.getKey());
                        sb.append(metricName).append("_total");
                        appendLabels(sb, labels);
                        sb.append(" ").append(formatDouble(valEntry.getValue())).append("\n");
                    }
                } else if (type == MetricType.GAUGE) {
                    sb.append("# HELP ").append(metricName).append(" ").append(metricName).append(" gauge\n");
                    sb.append("# TYPE ").append(metricName).append(" gauge\n");
                    Map<String, Double> values = repository.getCurrentValues(metricName, type);
                    for (Map.Entry<String, Double> valEntry : values.entrySet()) {
                        Map<String, String> labels = repository.parseLabelKey(valEntry.getKey());
                        sb.append(metricName);
                        appendLabels(sb, labels);
                        sb.append(" ").append(formatDouble(valEntry.getValue())).append("\n");
                    }
                } else if (type == MetricType.HISTOGRAM) {
                    sb.append("# HELP ").append(metricName).append("_seconds ").append(metricName).append(" histogram\n");
                    sb.append("# TYPE ").append(metricName).append("_seconds histogram\n");
                    Map<String, List<Double>> buckets = repository.getHistogramBuckets(metricName);
                    List<Double> bounds = properties.getBuckets().getHistogram();
                    
                    for (Map.Entry<String, List<Double>> bucketEntry : buckets.entrySet()) {
                        Map<String, String> labels = repository.parseLabelKey(bucketEntry.getKey());
                        List<Double> counts = bucketEntry.getValue();
                        
                        double cumulativeCount = 0;
                        for (int i = 0; i < bounds.size(); i++) {
                            cumulativeCount += counts.get(i);
                            Map<String, String> bucketLabels = new HashMap<>(labels);
                            bucketLabels.put("le", formatDouble(bounds.get(i)));
                            sb.append(metricName).append("_seconds_bucket");
                            appendLabels(sb, bucketLabels);
                            sb.append(" ").append(formatDouble(cumulativeCount)).append("\n");
                        }
                        cumulativeCount += counts.get(bounds.size());
                        Map<String, String> infLabels = new HashMap<>(labels);
                        infLabels.put("le", "+Inf");
                        sb.append(metricName).append("_seconds_bucket");
                        appendLabels(sb, infLabels);
                        sb.append(" ").append(formatDouble(cumulativeCount)).append("\n");
                        
                        sb.append(metricName).append("_seconds_sum");
                        appendLabels(sb, labels);
                        sb.append(" ").append(formatDouble(counts.get(bounds.size() + 1))).append("\n");
                        
                        sb.append(metricName).append("_seconds_count");
                        appendLabels(sb, labels);
                        sb.append(" ").append(formatDouble(cumulativeCount)).append("\n");
                    }
                }
            }
        }

        return sb.toString();
    }

    public List<QueryResponse> queryMetrics(String metricName, String type, Map<String, String> labels,
                                            Instant startTime, Instant endTime) {
        MetricType metricType = parseType(type);
        if (metricType == null) {
            throw new IllegalArgumentException("Invalid metric type: " + type);
        }

        List<DataPoint> rawData = repository.queryRawData(metricName, metricType, labels, startTime, endTime);
        
        if (rawData.isEmpty()) {
            return Collections.emptyList();
        }

        Map<String, List<DataPoint>> grouped = rawData.stream()
            .collect(Collectors.groupingBy(dp -> getLabelKey(dp.getLabels())));

        List<QueryResponse> results = new ArrayList<>();
        for (Map.Entry<String, List<DataPoint>> entry : grouped.entrySet()) {
            List<Double> values = entry.getValue().stream()
                .map(DataPoint::getValue)
                .sorted()
                .collect(Collectors.toList());

            QueryResponse resp = new QueryResponse();
            resp.setMetricName(metricName);
            resp.setLabels(repository.parseLabelKey(entry.getKey()));
            resp.setCount((long) values.size());
            resp.setAverage(values.stream().mapToDouble(Double::doubleValue).average().orElse(0.0));
            
            if (!values.isEmpty()) {
                resp.setP50(percentile(values, 0.5));
                resp.setP95(percentile(values, 0.95));
                resp.setP99(percentile(values, 0.99));
            }
            
            results.add(resp);
        }

        return results;
    }

    private MetricType parseType(String type) {
        if (type == null) return null;
        try {
            return MetricType.valueOf(type.toUpperCase());
        } catch (IllegalArgumentException e) {
            return null;
        }
    }

    private double percentile(List<Double> sortedValues, double p) {
        if (sortedValues.isEmpty()) return 0.0;
        int n = sortedValues.size();
        if (n == 1) return sortedValues.get(0);
        
        double index = (n - 1) * p;
        int lower = (int) Math.floor(index);
        int upper = (int) Math.ceil(index);
        
        if (lower == upper) {
            return sortedValues.get(lower);
        }
        
        double fraction = index - lower;
        return sortedValues.get(lower) * (1 - fraction) + sortedValues.get(upper) * fraction;
    }

    private void appendLabels(StringBuilder sb, Map<String, String> labels) {
        if (labels != null && !labels.isEmpty()) {
            sb.append("{");
            String joined = labels.entrySet().stream()
                .map(e -> e.getKey() + "=\"" + escapeLabelValue(e.getValue()) + "\"")
                .collect(Collectors.joining(","));
            sb.append(joined);
            sb.append("}");
        }
    }

    private String escapeLabelValue(String value) {
        if (value == null) return "";
        return value.replace("\\", "\\\\")
                    .replace("\"", "\\\"")
                    .replace("\n", "\\n");
    }

    private String formatDouble(double value) {
        if (value == (long) value) {
            return String.format("%d", (long) value);
        }
        return String.format("%f", value);
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
