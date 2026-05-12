package com.loadbalancer.strategy;

import com.loadbalancer.model.Instance;

import java.util.List;

public interface LoadBalanceStrategy {
    Instance select(List<Instance> availableInstances);
    String getName();
}
