package com.example.registry.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "registry")
public class RegistryProperties {
    
    private Heartbeat heartbeat = new Heartbeat();
    private GracefulShutdown gracefulShutdown = new GracefulShutdown();
    private Discovery discovery = new Discovery();
    private Notification notification = new Notification();
    
    @Data
    public static class Heartbeat {
        private long interval = 10000;
        private int unhealthyThreshold = 3;
        private int deregisterThreshold = 6;
        private long recoveryWindow = 60000;
    }
    
    @Data
    public static class GracefulShutdown {
        private long waitTime = 10000;
    }
    
    @Data
    public static class Discovery {
        private int pageSize = 20;
        private int paginationThreshold = 100;
    }
    
    @Data
    public static class Notification {
        private int maxRetries = 5;
    }
}
