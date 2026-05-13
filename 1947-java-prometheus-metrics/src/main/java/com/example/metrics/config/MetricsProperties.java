package com.example.metrics.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

import java.util.List;

@Data
@Configuration
@ConfigurationProperties(prefix = "metrics")
public class MetricsProperties {
    private Retention retention = new Retention();
    private Labels labels = new Labels();
    private Buckets buckets = new Buckets();

    @Data
    public static class Retention {
        private int rawDataHours = 1;
    }

    @Data
    public static class Labels {
        private int maxCombinations = 100;
    }

    @Data
    public static class Buckets {
        private List<Double> histogram = List.of(100.0, 500.0, 1000.0, 5000.0);
    }
}
