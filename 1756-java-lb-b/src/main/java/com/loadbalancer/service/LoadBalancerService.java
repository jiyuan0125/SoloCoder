package com.loadbalancer.service;

import com.loadbalancer.model.Instance;
import com.loadbalancer.registry.InstanceRegistry;
import com.loadbalancer.strategy.LoadBalanceStrategy;
import com.loadbalancer.strategy.LoadBalanceStrategyFactory;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class LoadBalancerService {
    private static final Logger logger = LoggerFactory.getLogger(LoadBalancerService.class);

    private final InstanceRegistry instanceRegistry;
    private final LoadBalanceStrategyFactory strategyFactory;
    private volatile String currentStrategy;

    public LoadBalancerService(InstanceRegistry instanceRegistry, LoadBalanceStrategyFactory strategyFactory) {
        this.instanceRegistry = instanceRegistry;
        this.strategyFactory = strategyFactory;
        this.currentStrategy = strategyFactory.getDefaultStrategy().getName();
    }

    public Instance getNextInstance() {
        return getNextInstance(currentStrategy);
    }

    public Instance getNextInstance(String strategyName) {
        List<Instance> availableInstances = instanceRegistry.getAvailableInstances();
        
        if (availableInstances.isEmpty()) {
            logger.warn("No available instances to handle request");
            return null;
        }

        LoadBalanceStrategy strategy = strategyFactory.getStrategy(strategyName);
        Instance selected = strategy.select(availableInstances);
        
        if (selected != null) {
            logger.debug("Selected instance: {} using strategy: {}", selected.getId(), strategy.getName());
        }
        
        return selected;
    }

    public void setStrategy(String strategyName) {
        if (strategyFactory.hasStrategy(strategyName)) {
            this.currentStrategy = strategyName;
            logger.info("Load balancer strategy changed to: {}", strategyName);
        } else {
            throw new IllegalArgumentException("Unknown strategy: " + strategyName);
        }
    }

    public String getCurrentStrategy() {
        return currentStrategy;
    }

    public void recordRequestSuccess(Instance instance) {
        instance.incrementTotalRequests();
        instance.incrementSuccessfulRequests();
    }

    public void recordRequestFailure(Instance instance) {
        instance.incrementTotalRequests();
        instance.incrementFailedRequests();
    }
}
