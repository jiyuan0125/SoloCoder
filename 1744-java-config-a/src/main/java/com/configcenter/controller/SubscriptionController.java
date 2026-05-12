package com.configcenter.controller;

import com.configcenter.dto.ApiResponse;
import com.configcenter.dto.ConfigChangeEvent;
import com.configcenter.dto.SubscriptionRequestDTO;
import com.configcenter.model.ClientSubscription;
import com.configcenter.service.SubscriptionService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/projects/{projectId}/subscriptions")
@RequiredArgsConstructor
public class SubscriptionController {

    private final SubscriptionService subscriptionService;

    @PostMapping
    public ApiResponse<ClientSubscription> subscribe(
            @PathVariable Long projectId,
            @Valid @RequestBody SubscriptionRequestDTO request) {
        ClientSubscription subscription = subscriptionService.subscribe(
                request.getInstanceId(),
                projectId,
                request.getEnvironment(),
                request.getLastKnownVersion()
        );
        return ApiResponse.success("Subscribed successfully", subscription);
    }

    @PostMapping("/heartbeat")
    public ApiResponse<Void> heartbeat(
            @PathVariable Long projectId,
            @Valid @RequestBody SubscriptionRequestDTO request) {
        subscriptionService.heartbeat(
                request.getInstanceId(),
                projectId,
                request.getEnvironment()
        );
        return ApiResponse.success("Heartbeat updated", null);
    }

    @PostMapping("/watch")
    public ApiResponse<ConfigChangeEvent> watch(
            @PathVariable Long projectId,
            @Valid @RequestBody SubscriptionRequestDTO request) {
        subscriptionService.subscribe(
                request.getInstanceId(),
                projectId,
                request.getEnvironment(),
                request.getLastKnownVersion()
        );
        
        ConfigChangeEvent event = subscriptionService.waitForChanges(
                request.getInstanceId(),
                projectId,
                request.getEnvironment(),
                request.getLastKnownVersion()
        );
        
        if (event == null) {
            return ApiResponse.success("No changes during timeout period", null);
        }
        return ApiResponse.success("Config changes detected", event);
    }

    @GetMapping
    public ApiResponse<List<ClientSubscription>> getSubscribers(
            @PathVariable Long projectId,
            @RequestParam String environment) {
        List<ClientSubscription> subscribers = subscriptionService.getSubscribers(projectId, environment);
        return ApiResponse.success(subscribers);
    }
}
