package com.example.loadbalancer.model;

public enum LoadBalanceStrategy {
    ROUND_ROBIN,
    WEIGHTED_ROUND_ROBIN,
    LEAST_CONN;

    public static LoadBalanceStrategy fromString(String value) {
        switch (value.toLowerCase()) {
            case "round_robin":
                return ROUND_ROBIN;
            case "weighted_round_robin":
                return WEIGHTED_ROUND_ROBIN;
            case "least_conn":
                return LEAST_CONN;
            default:
                throw new IllegalArgumentException("Unknown strategy: " + value);
        }
    }
}
