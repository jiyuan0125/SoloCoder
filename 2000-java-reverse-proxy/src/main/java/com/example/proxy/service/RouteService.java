package com.example.proxy.service;

import com.example.proxy.config.ProxyProperties;
import com.example.proxy.model.Backend;
import com.example.proxy.model.Route;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Slf4j
@Service
public class RouteService {
    
    private final ProxyProperties properties;
    private final Map<String, Route> routes = new ConcurrentHashMap<>();
    private final Map<String, Route> exactRoutes = new ConcurrentHashMap<>();
    private final TreeMap<String, Route> prefixRoutes = new TreeMap<>(Comparator.reverseOrder());
    
    public RouteService(ProxyProperties properties) {
        this.properties = properties;
    }
    
    @PostConstruct
    public void init() {
        log.info("Initializing routes from configuration...");
        for (ProxyProperties.RouteConfig config : properties.getRoutes()) {
            Route route = convertToRoute(config);
            addRoute(route);
        }
        log.info("Initialized {} routes", routes.size());
    }
    
    private Route convertToRoute(ProxyProperties.RouteConfig config) {
        List<Backend> backends = config.getBackends().stream()
                .map(this::convertToBackend)
                .collect(Collectors.toList());
        
        return Route.builder()
                .id(UUID.randomUUID().toString())
                .path(config.getPath())
                .type(Route.RouteType.valueOf(config.getType().toUpperCase()))
                .backends(backends)
                .build();
    }
    
    private Backend convertToBackend(ProxyProperties.BackendConfig config) {
        return Backend.builder()
                .id(UUID.randomUUID().toString())
                .url(config.getUrl())
                .weight(config.getWeight())
                .status(Backend.BackendStatus.HEALTHY)
                .consecutiveFailures(0)
                .consecutiveSuccesses(0)
                .consecutive5xxCount(0)
                .build();
    }
    
    public Route matchRoute(String requestPath) {
        log.debug("Matching route for path: {}", requestPath);
        
        Route exactMatch = exactRoutes.get(requestPath);
        if (exactMatch != null) {
            log.debug("Exact match found: {}", exactMatch.getPath());
            return exactMatch;
        }
        
        for (Map.Entry<String, Route> entry : prefixRoutes.entrySet()) {
            String prefix = entry.getKey();
            if (requestPath.equals(prefix) || requestPath.startsWith(prefix + "/")) {
                log.debug("Prefix match found: {}", prefix);
                return entry.getValue();
            }
        }
        
        log.debug("No route matched for path: {}", requestPath);
        return null;
    }
    
    public void addRoute(Route route) {
        String routeKey = generateRouteKey(route.getPath(), route.getType());
        routes.put(routeKey, route);
        
        if (route.getType() == Route.RouteType.EXACT) {
            exactRoutes.put(route.getPath(), route);
            log.info("Added exact route: {}", route.getPath());
        } else {
            prefixRoutes.put(route.getPath(), route);
            log.info("Added prefix route: {}", route.getPath());
        }
    }
    
    public boolean removeRoute(String path, Route.RouteType type) {
        String routeKey = generateRouteKey(path, type);
        Route route = routes.remove(routeKey);
        
        if (route != null) {
            if (type == Route.RouteType.EXACT) {
                exactRoutes.remove(path);
            } else {
                prefixRoutes.remove(path);
            }
            log.info("Removed route: {} ({})", path, type);
            return true;
        }
        return false;
    }
    
    public List<Route> getAllRoutes() {
        return new ArrayList<>(routes.values());
    }
    
    public Optional<Route> getRoute(String path, Route.RouteType type) {
        return Optional.ofNullable(routes.get(generateRouteKey(path, type)));
    }
    
    private String generateRouteKey(String path, Route.RouteType type) {
        return type + ":" + path;
    }
    
    public void addBackendToRoute(String routePath, Route.RouteType routeType, Backend backend) {
        Optional<Route> routeOpt = getRoute(routePath, routeType);
        routeOpt.ifPresent(route -> {
            route.getBackends().add(backend);
            log.info("Added backend {} to route {}", backend.getUrl(), routePath);
        });
    }
    
    public boolean removeBackendFromRoute(String routePath, Route.RouteType routeType, String backendUrl) {
        Optional<Route> routeOpt = getRoute(routePath, routeType);
        if (routeOpt.isPresent()) {
            Route route = routeOpt.get();
            boolean removed = route.getBackends().removeIf(b -> b.getUrl().equals(backendUrl));
            if (removed) {
                log.info("Removed backend {} from route {}", backendUrl, routePath);
            }
            return removed;
        }
        return false;
    }
}
