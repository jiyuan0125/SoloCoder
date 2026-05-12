package com.loadbalancer.service;

import com.loadbalancer.model.LoadBalancingStrategy;
import com.loadbalancer.model.StrategyState;
import com.loadbalancer.strategy.LeastConnectionsStrategy;
import com.loadbalancer.strategy.LoadBalancerStrategy;
import com.loadbalancer.strategy.WeightedRoundRobinStrategy;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.util.concurrent.atomic.AtomicReference;

@Service
public class StrategyManager {

    private final WeightedRoundRobinStrategy weightedRoundRobinStrategy;
    private final LeastConnectionsStrategy leastConnectionsStrategy;

    private final AtomicReference<LoadBalancingStrategy> currentStrategy = new AtomicReference<>();
    private final AtomicReference<LoadBalancingStrategy> pendingStrategy = new AtomicReference<>();
    private final AtomicReference<StrategyState> currentState = new AtomicReference<>(StrategyState.ACTIVE);

    public StrategyManager(WeightedRoundRobinStrategy weightedRoundRobinStrategy,
                        LeastConnectionsStrategy leastConnectionsStrategy,
                        @Value("${loadbalancer.strategy:WEIGHTED_ROUND_ROBIN}") String initialStrategy) {
        this.weightedRoundRobinStrategy = weightedRoundRobinStrategy;
        this.leastConnectionsStrategy = leastConnectionsStrategy;
        this.currentStrategy.set(LoadBalancingStrategy.valueOf(initialStrategy));
    }

    public boolean configureStrategy(LoadBalancingStrategy newStrategy) {
        StrategyState current = currentState.get();
        
        if (current != StrategyState.ACTIVE) {
            return false;
        }

        if (newStrategy == currentStrategy.get()) {
            return false;
        }

        pendingStrategy.set(newStrategy);
        currentState.set(StrategyState.CONFIGURING);
        return true;
    }

    public boolean startActivation() {
        StrategyState current = currentState.get();
        
        if (current != StrategyState.CONFIGURING) {
            return false;
        }

        if (pendingStrategy.get() == null) {
            return false;
        }

        currentState.set(StrategyState.ACTIVATING);
        return true;
    }

    public boolean commit() {
        StrategyState current = currentState.get();
        
        if (current != StrategyState.ACTIVATING) {
            return false;
        }

        LoadBalancingStrategy pending = pendingStrategy.get();
        if (pending == null) {
            return false;
        }

        currentStrategy.set(pending);
        pendingStrategy.set(null);
        currentState.set(StrategyState.ACTIVE);
        return true;
    }

    public boolean cancelConfiguration() {
        StrategyState current = currentState.get();
        
        if (current == StrategyState.ACTIVE) {
            return false;
        }

        pendingStrategy.set(null);
        currentState.set(StrategyState.ACTIVE);
        return true;
    }

    public LoadBalancingStrategy getCurrentStrategy() {
        return currentStrategy.get();
    }

    public LoadBalancingStrategy getPendingStrategy() {
        return pendingStrategy.get();
    }

    public StrategyState getCurrentState() {
        return currentState.get();
    }

    public LoadBalancerStrategy getCurrentStrategyImplementation() {
        return getStrategyImplementation(currentStrategy.get());
    }

    private LoadBalancerStrategy getStrategyImplementation(LoadBalancingStrategy strategy) {
        return switch (strategy) {
            case WEIGHTED_ROUND_ROBIN -> weightedRoundRobinStrategy;
            case LEAST_CONNECTIONS -> leastConnectionsStrategy;
        };
    }
}
