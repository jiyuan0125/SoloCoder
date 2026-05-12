package com.loadbalancer.strategy;

import com.loadbalancer.config.LoadBalancerConfig;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Component
public class LoadBalanceStrategyFactory {
    private final Map<String, LoadBalanceStrategy> strategies = new HashMap<>();
    private final String defaultStrategyName;

    public LoadBalanceStrategyFactory(List<LoadBalanceStrategy> strategyList, LoadBalancerConfig config) {
        for (LoadBalanceStrategy strategy : strategyList) {
            strategies.put(strategy.getName(), strategy);
        }
        this.defaultStrategyName = config.getDefaultStrategy();
    }

    public LoadBalanceStrategy getStrategy(String name) {
        return strategies.getOrDefault(name, getDefaultStrategy());
    }

    public LoadBalanceStrategy getDefaultStrategy() {
        return strategies.get(defaultStrategyName);
    }

    public boolean hasStrategy(String name) {
        return strategies.containsKey(name);
    }
}
