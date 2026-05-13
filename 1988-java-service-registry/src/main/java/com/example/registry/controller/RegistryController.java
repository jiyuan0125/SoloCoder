package com.example.registry.controller;

import com.example.registry.dto.*;
import com.example.registry.model.ServiceInstance;
import com.example.registry.model.Subscriber;
import com.example.registry.service.GracefulShutdownService;
import com.example.registry.service.HeartbeatService;
import com.example.registry.service.ServiceRegistryService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/api")
public class RegistryController {

    private final ServiceRegistryService registryService;
    private final HeartbeatService heartbeatService;
    private final GracefulShutdownService gracefulShutdownService;

    public RegistryController(ServiceRegistryService registryService,
                              HeartbeatService heartbeatService,
                              GracefulShutdownService gracefulShutdownService) {
        this.registryService = registryService;
        this.heartbeatService = heartbeatService;
        this.gracefulShutdownService = gracefulShutdownService;
    }

    @PostMapping("/register")
    public ResponseEntity<RegisterResponse> register(@Valid @RequestBody RegisterRequest request) {
        ServiceInstance instance = registryService.register(
            request.getServiceName(),
            request.getIp(),
            request.getPort(),
            request.getMetadata()
        );

        return ResponseEntity.ok(RegisterResponse.builder()
            .instanceId(instance.getInstanceId())
            .leaseDuration(instance.getLeaseDuration())
            .message("Registered successfully")
            .build());
    }

    @PutMapping("/heartbeat/{serviceName}/{instanceId}")
    public ResponseEntity<HeartbeatResponse> heartbeat(
            @PathVariable String serviceName,
            @PathVariable String instanceId) {
        
        boolean success = heartbeatService.heartbeat(serviceName, instanceId);
        
        if (success) {
            return ResponseEntity.ok(HeartbeatResponse.builder()
                .success(true)
                .message("Heartbeat accepted")
                .build());
        } else {
            return ResponseEntity.status(404).body(HeartbeatResponse.builder()
                .success(false)
                .message("Instance not found or deregistered")
                .build());
        }
    }

    @GetMapping("/discover/{serviceName}")
    public ResponseEntity<DiscoveryResponse> discover(
            @PathVariable String serviceName,
            @RequestParam(required = false) Map<String, String> allParams,
            @RequestParam(required = false) Integer page) {
        
        Map<String, String> metadataFilter = new HashMap<>();
        if (allParams != null) {
            allParams.forEach((key, value) -> {
                if (!"page".equals(key)) {
                    metadataFilter.put(key, value);
                }
            });
        }

        DiscoveryResponse response = registryService.discover(serviceName, metadataFilter, page);
        return ResponseEntity.ok(response);
    }

    @PutMapping("/metadata/{serviceName}/{instanceId}")
    public ResponseEntity<Void> updateMetadata(
            @PathVariable String serviceName,
            @PathVariable String instanceId,
            @RequestBody MetadataUpdateRequest request) {
        
        boolean success = registryService.updateMetadata(serviceName, instanceId, request.getMetadata());
        
        if (success) {
            return ResponseEntity.ok().build();
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    @PostMapping("/deregister/{serviceName}/{instanceId}")
    public ResponseEntity<Void> deregister(
            @PathVariable String serviceName,
            @PathVariable String instanceId) {
        
        boolean success = gracefulShutdownService.initiateGracefulShutdown(serviceName, instanceId);
        
        if (success) {
            return ResponseEntity.ok().build();
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    @PostMapping("/cancel-deregister/{serviceName}/{instanceId}")
    public ResponseEntity<Void> cancelDeregister(
            @PathVariable String serviceName,
            @PathVariable String instanceId) {
        
        boolean success = gracefulShutdownService.cancelGracefulShutdown(serviceName, instanceId);
        
        if (success) {
            return ResponseEntity.ok().build();
        } else {
            return ResponseEntity.notFound().build();
        }
    }

    @PostMapping("/subscribe")
    public ResponseEntity<Subscriber> subscribe(@Valid @RequestBody SubscribeRequest request) {
        Subscriber subscriber = registryService.subscribe(
            request.getServiceName(),
            request.getCallbackUrl()
        );
        return ResponseEntity.ok(subscriber);
    }

    @PostMapping("/unsubscribe")
    public ResponseEntity<Void> unsubscribe(@Valid @RequestBody SubscribeRequest request) {
        registryService.unsubscribe(request.getServiceName(), request.getCallbackUrl());
        return ResponseEntity.ok().build();
    }
}
