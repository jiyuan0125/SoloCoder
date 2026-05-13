package com.example.proxy.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

@Data
@Component
@ConfigurationProperties(prefix = "proxy")
public class ProxyProperties {
    
    private int requestTimeout = 5000;
    private int connectTimeout = 5000;
    private String loadBalanceStrategy = "round-robin";
    private long failureRecoveryTime = 60000;
    private int consecutive5xxThreshold = 3;
    
    private HealthCheckConfig healthCheck = new HealthCheckConfig();
    private AccessLogConfig accessLog = new AccessLogConfig();
    private List<RouteConfig> routes = new ArrayList<>();
    private ManagementConfig management = new ManagementConfig();
    
    @Data
    public static class HealthCheckConfig {
        private long interval = 10000;
        private String path = "/health";
        private int consecutiveFailuresThreshold = 3;
        private int consecutiveSuccessesRecovery = 2;
    }
    
    @Data
    public static class AccessLogConfig {
        private int maxEntries = 10000;
    }
    
    @Data
    public static class RouteConfig {
        private String path;
        private String type = "prefix";
        private List<BackendConfig> backends = new ArrayList<>();
    }
    
    @Data
    public static class BackendConfig {
        private String url;
        private int weight = 1;
    }
    
    @Data
    public static class ManagementConfig {
        private String basePath = "/admin";
    }
}
