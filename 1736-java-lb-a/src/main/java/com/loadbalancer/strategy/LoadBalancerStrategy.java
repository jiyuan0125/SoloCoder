package com.loadbalancer.strategy;

import com.loadbalancer.model.BackendInstance;

import java.util.List;

public interface LoadBalancerStrategy {
    BackendInstance selectInstance(List<BackendInstance> instances);
}
