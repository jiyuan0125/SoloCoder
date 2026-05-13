package com.example.gateway.routing;

import java.util.Collection;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;

public class RouteStore {
    private final Map<String, RouteRule> routes = new ConcurrentHashMap<>();
    private final AtomicInteger counter = new AtomicInteger(0);

    public void addRoute(RouteRule rule) {
        routes.put(rule.getVersion(), rule);
    }

    public RouteRule getRoute(String version) {
        return routes.get(version);
    }

    public Collection<RouteRule> getAllRoutes() {
        return routes.values();
    }

    public boolean removeRoute(String version) {
        return routes.remove(version) != null;
    }

    public boolean hasRoute(String version) {
        return routes.containsKey(version);
    }

    public String pickBackend(String version) {
        RouteRule rule = routes.get(version);
        if (rule == null || rule.getBackends().isEmpty()) {
            return null;
        }
        List<String> backends = rule.getBackends();
        int idx = Math.abs(counter.getAndIncrement()) % backends.size();
        return backends.get(idx);
    }
}
