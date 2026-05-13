package com.example.cachemiddleware.config;

import com.example.cachemiddleware.model.EvictionPolicy;
import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "cache")
public class CacheProperties {

    private DefaultConfig defaultConfig = new DefaultConfig();
    private int maxValueSizeBytes = 1048576;

    @Data
    public static class DefaultConfig {
        private long ttlSeconds = 300;
        private int maxCapacity = 1000;
        private EvictionPolicy evictionPolicy = EvictionPolicy.LRU;
    }
}
