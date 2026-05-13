package com.example.registry.service;

import com.example.registry.dto.SubscribeRequest;
import com.example.registry.model.Subscription;
import com.example.registry.store.SubscriptionStore;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class SubscriptionService {
    private final SubscriptionStore subscriptionStore;

    public Subscription subscribe(SubscribeRequest request) {
        Subscription subscription = Subscription.builder()
                .id(UUID.randomUUID().toString())
                .serviceName(request.getServiceName())
                .callbackUrl(request.getCallbackUrl())
                .subscribeTime(Instant.now())
                .build();

        Subscription saved = subscriptionStore.save(subscription);
        log.info("Subscription created: {} for service {}", saved.getId(), saved.getServiceName());
        return saved;
    }
}
