package com.example.registry.store;

import com.example.registry.model.Subscription;
import org.springframework.stereotype.Component;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Component
public class SubscriptionStore {
    private final Map<String, Subscription> subscriptionsById = new ConcurrentHashMap<>();
    private final Map<String, Set<String>> subscriptionIdsByService = new ConcurrentHashMap<>();

    public Subscription save(Subscription subscription) {
        subscriptionsById.put(subscription.getId(), subscription);
        subscriptionIdsByService.computeIfAbsent(subscription.getServiceName(), k -> ConcurrentHashMap.newKeySet())
                .add(subscription.getId());
        return subscription;
    }

    public List<Subscription> findByServiceName(String serviceName) {
        Set<String> ids = subscriptionIdsByService.get(serviceName);
        if (ids == null || ids.isEmpty()) {
            return Collections.emptyList();
        }
        List<Subscription> result = new ArrayList<>();
        for (String id : ids) {
            Subscription sub = subscriptionsById.get(id);
            if (sub != null) {
                result.add(sub);
            }
        }
        return result;
    }

    public void remove(String id) {
        Subscription sub = subscriptionsById.remove(id);
        if (sub != null) {
            Set<String> ids = subscriptionIdsByService.get(sub.getServiceName());
            if (ids != null) {
                ids.remove(id);
                if (ids.isEmpty()) {
                    subscriptionIdsByService.remove(sub.getServiceName());
                }
            }
        }
    }
}
