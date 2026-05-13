package com.loadbalancer.dto;

import lombok.Data;
import lombok.AllArgsConstructor;
import java.util.List;

@Data
@AllArgsConstructor
public class DashboardResponse {
    private List<NodeDashboard> nodes;

    @Data
    @AllArgsConstructor
    public static class NodeDashboard {
        private String id;
        private String address;
        private int currentWeight;
        private String status;
        private int activeConnections;
        private List<HealthCheckRecord> recentChecks;
    }

    @Data
    @AllArgsConstructor
    public static class HealthCheckRecord {
        private String timestamp;
        private String result;
    }
}
