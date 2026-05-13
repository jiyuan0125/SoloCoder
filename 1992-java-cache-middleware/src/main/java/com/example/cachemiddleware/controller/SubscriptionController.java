package com.example.cachemiddleware.controller;

import com.example.cachemiddleware.model.SubscribeRequest;
import com.example.cachemiddleware.notify.SubscriptionManager;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/api/namespaces/{namespace}/subscriptions")
public class SubscriptionController {

    private final SubscriptionManager subscriptionManager;

    public SubscriptionController(SubscriptionManager subscriptionManager) {
        this.subscriptionManager = subscriptionManager;
    }

    @PostMapping
    public ResponseEntity<Map<String, Object>> subscribe(
            @PathVariable String namespace,
            @Valid @RequestBody SubscribeRequest request) {
        subscriptionManager.subscribe(namespace, request.getCallbackUrl());
        Map<String, Object> result = new HashMap<>();
        result.put("success", true);
        return ResponseEntity.ok(result);
    }

    @DeleteMapping
    public ResponseEntity<Map<String, Object>> unsubscribe(
            @PathVariable String namespace,
            @RequestBody SubscribeRequest request) {
        subscriptionManager.unsubscribe(namespace, request.getCallbackUrl());
        Map<String, Object> result = new HashMap<>();
        result.put("success", true);
        return ResponseEntity.ok(result);
    }
}
