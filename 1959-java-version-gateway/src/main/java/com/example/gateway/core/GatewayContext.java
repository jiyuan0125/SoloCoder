package com.example.gateway.core;

import com.example.gateway.client.BackendClient;
import com.example.gateway.compatibility.CompatibilityService;
import com.example.gateway.config.GatewayConfig;
import com.example.gateway.routing.RouteStore;
import com.example.gateway.stats.StatsStore;

public class GatewayContext {
    private final GatewayConfig config;
    private final RouteStore routeStore;
    private final StatsStore statsStore;
    private final CompatibilityService compatibilityService;
    private final BackendClient backendClient;

    public GatewayContext(GatewayConfig config) {
        this.config = config;
        this.routeStore = new RouteStore();
        this.statsStore = new StatsStore();
        this.compatibilityService = new CompatibilityService();
        this.backendClient = new BackendClient();
    }

    public GatewayConfig getConfig() {
        return config;
    }

    public RouteStore getRouteStore() {
        return routeStore;
    }

    public StatsStore getStatsStore() {
        return statsStore;
    }

    public CompatibilityService getCompatibilityService() {
        return compatibilityService;
    }

    public BackendClient getBackendClient() {
        return backendClient;
    }

    public void shutdown() {
        backendClient.shutdown();
    }
}
