package com.example.loadbalancer.strategy;

import com.example.loadbalancer.model.Node;

public interface LoadBalanceStrategy {
    String getName();
    Node select();
}
