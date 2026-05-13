package com.example.proxy.service;

import com.example.proxy.config.ProxyProperties;
import com.example.proxy.loadbalance.LoadBalanceStrategy;
import com.example.proxy.model.Backend;
import com.example.proxy.model.Route;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

@Slf4j
@Service
public class LoadBalanceService {
    
    private final ProxyProperties properties;
    private final List<LoadBalanceStrategy> strategies;
    private Map<String, LoadBalanceStrategy> strategyMap;
    private LoadBalanceStrategy defaultStrategy;
    
    public LoadBalanceService(ProxyProperties properties, List<LoadBalanceStrategy> strategies) {
        this.properties = properties;
        this.strategies = strategies;
    }
    
    @PostConstruct
    public void init() {
        this.strategyMap = strategies.stream()
                .collect(Collectors.toMap(LoadBalanceStrategy::getName, s -> s));
        
        String defaultStrategyName = properties.getLoadBalanceStrategy();
        this.defaultStrategy = strategyMap.get(defaultStrategyName);
        
        if (defaultStrategy == null) {
            log.warn("Default strategy '{}' not found, using first available: {}", 
                    defaultStrategyName, strategies.get(0).getName());
            this.defaultStrategy = strategies.get(0);
        }
        
        log.info("LoadBalanceService initialized with default strategy: {}", defaultStrategy.getName());
    }
    
    public Backend selectBackend(Route route) {
        return selectBackend(route, null);
    }
    
    public Backend selectBackend(Route route, String strategyName) {
        if (route == null || route.getBackends() == null || route.getBackends().isEmpty()) {
            log.warn("No backends available for route: {}", route != null ? route.getPath() : "null");
            return null;
        }
        
        List<Backend> availableBackends = route.getBackends().stream()
                .filter(Backend::isAvailable)
                .collect(Collectors.toList());
        
        if (availableBackends.isEmpty()) {
            log.warn("No healthy backends available for route: {}", route.getPath());
            return null;
        }
        
        LoadBalanceStrategy strategy = strategyName != null ? 
                strategyMap.get(strategyName) : defaultStrategy;
        
        if (strategy == null) {
            log.warn("Strategy '{}' not found, using default: {}", strategyName, defaultStrategy.getName());
            strategy = defaultStrategy;
        }
        
        Backend selected = strategy.select(availableBackends);
        log.debug("Selected backend {} for route {}", 
                selected != null ? selected.getUrl() : "null", route.getPath());
        
        return selected;
    }
    
    public Optional<LoadBalanceStrategy> getStrategy(String name) {
        return Optional.ofNullable(strategyMap.get(name));
    }
    
    public List<String> getAvailableStrategies() {
        return strategies.stream()
                .map(LoadBalanceStrategy::getName)
                .collect(Collectors.toList());
    }
}
