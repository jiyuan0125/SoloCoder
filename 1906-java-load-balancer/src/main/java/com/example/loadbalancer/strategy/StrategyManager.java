package com.example.loadbalancer.strategy;

import com.example.loadbalancer.model.LoadBalanceStrategy;
import com.example.loadbalancer.model.Node;
import org.springframework.stereotype.Component;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicReference;

@Component
public class StrategyManager {

    private final Map<String, com.example.loadbalancer.strategy.LoadBalanceStrategy> strategies = new ConcurrentHashMap<>();
    private final AtomicReference<com.example.loadbalancer.strategy.LoadBalanceStrategy> currentStrategy = new AtomicReference<>();

    public StrategyManager(RoundRobinStrategy roundRobin,
                           WeightedRoundRobinStrategy weightedRoundRobin,
                           LeastConnectionStrategy leastConnection) {
        strategies.put("round_robin", roundRobin);
        strategies.put("weighted_round_robin", weightedRoundRobin);
        strategies.put("least_conn", leastConnection);
        this.currentStrategy.set(roundRobin);
    }

    public void setInitialStrategy(String strategyName) {
        com.example.loadbalancer.strategy.LoadBalanceStrategy strategy = strategies.get(strategyName);
        if (strategy != null) {
            this.currentStrategy.set(strategy);
        }
    }

    public boolean switchStrategy(String strategyName) {
        com.example.loadbalancer.strategy.LoadBalanceStrategy strategy = strategies.get(strategyName);
        if (strategy == null) {
            return false;
        }
        this.currentStrategy.set(strategy);
        return true;
    }

    public boolean switchStrategy(LoadBalanceStrategy strategy) {
        String name;
        switch (strategy) {
            case ROUND_ROBIN:
                name = "round_robin";
                break;
            case WEIGHTED_ROUND_ROBIN:
                name = "weighted_round_robin";
                break;
            case LEAST_CONN:
                name = "least_conn";
                break;
            default:
                return false;
        }
        return switchStrategy(name);
    }

    public String getCurrentStrategyName() {
        com.example.loadbalancer.strategy.LoadBalanceStrategy strategy = currentStrategy.get();
        return strategy != null ? strategy.getName() : "round_robin";
    }

    public Node selectNode() {
        com.example.loadbalancer.strategy.LoadBalanceStrategy strategy = currentStrategy.get();
        return strategy != null ? strategy.select() : null;
    }
}
