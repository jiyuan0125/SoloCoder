package com.breaker.controller;

import com.breaker.core.BreakerManager;
import com.breaker.core.SubscriptionManager;
import com.breaker.dto.SubscriptionRequest;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;

@RestController
@RequestMapping("/subscriptions")
@Validated
public class SubscriptionController {

    private final SubscriptionManager subscriptionManager;
    private final BreakerManager breakerManager;

    public SubscriptionController(SubscriptionManager subscriptionManager, BreakerManager breakerManager) {
        this.subscriptionManager = subscriptionManager;
        this.breakerManager = breakerManager;
        this.breakerManager.setStateChangeListener(breaker -> {
            if (breaker.getLastStateChange() != null) {
                this.subscriptionManager.notifyStateChange(
                        breaker.getServiceName(),
                        breaker.getLastStateChange()
                );
            }
        });
    }

    @PostMapping
    public ResponseEntity<Void> subscribe(@Valid @RequestBody SubscriptionRequest request) {
        subscriptionManager.addSubscription(request.getServiceName(), request.getCallbackUrl());
        return ResponseEntity.ok().build();
    }
}
