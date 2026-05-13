package com.example.proxy.loadbalance;

import com.example.proxy.model.Backend;

import java.util.List;

public interface LoadBalanceStrategy {
    Backend select(List<Backend> availableBackends);
    String getName();
}
