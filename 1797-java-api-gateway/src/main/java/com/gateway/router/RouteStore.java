package com.gateway.router;

import com.gateway.model.RouteRule;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.*;
import java.util.concurrent.atomic.AtomicReference;

@Slf4j
@Component
public class RouteStore {

    private final AtomicReference<List<RouteRule>> routes = new AtomicReference<>(new ArrayList<>());

    public void addRoute(RouteRule route) {
        List<RouteRule> newRoutes = new ArrayList<>(routes.get());
        
        newRoutes.removeIf(r -> r.getId().equals(route.getId()));
        newRoutes.add(route);
        
        routes.set(Collections.unmodifiableList(newRoutes));
        log.info("已添加/更新路由规则: {} -> {}", route.getPathPrefix(), route.getId());
    }

    public boolean deleteRoute(String routeId) {
        List<RouteRule> existing = routes.get();
        List<RouteRule> newRoutes = new ArrayList<>(existing.size());
        
        boolean removed = false;
        for (RouteRule route : existing) {
            if (!route.getId().equals(routeId)) {
                newRoutes.add(route);
            } else {
                removed = true;
            }
        }
        
        if (removed) {
            routes.set(Collections.unmodifiableList(newRoutes));
            log.info("已删除路由规则: {}", routeId);
        }
        return removed;
    }

    public RouteRule getRoute(String routeId) {
        for (RouteRule route : routes.get()) {
            if (route.getId().equals(routeId)) {
                return route;
            }
        }
        return null;
    }

    public List<RouteRule> getAllRoutes() {
        return routes.get();
    }

    public void updateRoutes(List<RouteRule> newRouteList) {
        routes.set(Collections.unmodifiableList(new ArrayList<>(newRouteList)));
        log.info("已批量更新 {} 条路由规则", newRouteList.size());
    }
}
