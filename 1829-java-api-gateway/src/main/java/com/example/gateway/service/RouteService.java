package com.example.gateway.service;

import com.example.gateway.model.BackendTarget;
import com.example.gateway.model.RouteRule;
import org.springframework.stereotype.Service;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicReference;

@Service
public class RouteService {

    private final AtomicReference<Map<String, RouteRule>> routesRef = new AtomicReference<>(new ConcurrentHashMap<>());
    private final Random random = new Random();

    public RouteRule matchRoute(String path) {
        Map<String, RouteRule> currentRoutes = routesRef.get();
        List<RouteRule> matchedRoutes = new ArrayList<>();

        for (RouteRule rule : currentRoutes.values()) {
            if (path.startsWith(rule.getPathPrefix())) {
                matchedRoutes.add(rule);
            }
        }

        if (matchedRoutes.isEmpty()) {
            return null;
        }

        if (matchedRoutes.size() == 1) {
            return matchedRoutes.get(0);
        }

        matchedRoutes.sort((r1, r2) -> {
            int prefixCompare = Integer.compare(r2.getPathPrefix().length(), r1.getPathPrefix().length());
            if (prefixCompare != 0) return prefixCompare;
            return Integer.compare(r2.getWeight(), r1.getWeight());
        });

        int totalWeight = matchedRoutes.stream().mapToInt(RouteRule::getWeight).sum();
        int randomWeight = random.nextInt(totalWeight) + 1;
        int currentWeight = 0;

        for (RouteRule rule : matchedRoutes) {
            currentWeight += rule.getWeight();
            if (randomWeight <= currentWeight) {
                return rule;
            }
        }

        return matchedRoutes.get(0);
    }

    public BackendTarget selectBackend(RouteRule route) {
        List<BackendTarget> backends = route.getBackends();
        if (backends == null || backends.isEmpty()) {
            return null;
        }

        if (backends.size() == 1) {
            return backends.get(0);
        }

        int totalWeight = backends.stream().mapToInt(BackendTarget::getWeight).sum();
        int randomWeight = random.nextInt(totalWeight) + 1;
        int currentWeight = 0;

        for (BackendTarget target : backends) {
            currentWeight += target.getWeight();
            if (randomWeight <= currentWeight) {
                return target;
            }
        }

        return backends.get(0);
    }

    public void addRoute(RouteRule route) {
        Map<String, RouteRule> newRoutes = new ConcurrentHashMap<>(routesRef.get());
        newRoutes.put(route.getId(), route);
        routesRef.set(newRoutes);
    }

    public boolean updateRoute(String id, RouteRule route) {
        Map<String, RouteRule> currentRoutes = routesRef.get();
        if (!currentRoutes.containsKey(id)) {
            return false;
        }
        Map<String, RouteRule> newRoutes = new ConcurrentHashMap<>(currentRoutes);
        route.setId(id);
        newRoutes.put(id, route);
        routesRef.set(newRoutes);
        return true;
    }

    public boolean deleteRoute(String id) {
        Map<String, RouteRule> currentRoutes = routesRef.get();
        if (!currentRoutes.containsKey(id)) {
            return false;
        }
        Map<String, RouteRule> newRoutes = new ConcurrentHashMap<>(currentRoutes);
        newRoutes.remove(id);
        routesRef.set(newRoutes);
        return true;
    }

    public RouteRule getRoute(String id) {
        return routesRef.get().get(id);
    }

    public Collection<RouteRule> getAllRoutes() {
        return new ArrayList<>(routesRef.get().values());
    }

    public void updateFilterChain(String routeId, List<String> filterNames) {
        Map<String, RouteRule> currentRoutes = routesRef.get();
        RouteRule route = currentRoutes.get(routeId);
        if (route != null) {
            Map<String, RouteRule> newRoutes = new ConcurrentHashMap<>(currentRoutes);
            RouteRule updatedRoute = new RouteRule(
                route.getId(),
                route.getPathPrefix(),
                route.getBackends(),
                filterNames,
                route.getWeight()
            );
            newRoutes.put(routeId, updatedRoute);
            routesRef.set(newRoutes);
        }
    }
}
