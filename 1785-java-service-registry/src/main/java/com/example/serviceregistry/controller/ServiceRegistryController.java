package com.example.serviceregistry.controller;

import com.example.serviceregistry.dto.HeartbeatRequest;
import com.example.serviceregistry.dto.RegisterRequest;
import com.example.serviceregistry.dto.SubscribeRequest;
import com.example.serviceregistry.model.ServiceInstance;
import com.example.serviceregistry.model.Subscriber;
import com.example.serviceregistry.service.ServiceRegistry;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequiredArgsConstructor
public class ServiceRegistryController {

    private final ServiceRegistry serviceRegistry;

    @PostMapping("/register")
    public ResponseEntity<ServiceInstance> register(@Valid @RequestBody RegisterRequest request) {
        ServiceInstance instance = serviceRegistry.register(request);
        return ResponseEntity.ok(instance);
    }

    @PostMapping("/heartbeat")
    public ResponseEntity<Void> heartbeat(@Valid @RequestBody HeartbeatRequest request) {
        return serviceRegistry.heartbeat(request.getServiceName(), request.getInstanceId())
                .map(instance -> ResponseEntity.ok().<Void>build())
                .orElse(ResponseEntity.notFound().build());
    }

    @DeleteMapping("/deregister/{serviceName}/{instanceId}")
    public ResponseEntity<Void> deregister(
            @PathVariable String serviceName,
            @PathVariable String instanceId) {
        serviceRegistry.deregister(serviceName, instanceId);
        return ResponseEntity.ok().build();
    }

    @GetMapping("/services")
    public ResponseEntity<List<ServiceInstance>> getAllServices(
            @RequestParam(required = false) Map<String, String> allParams) {
        Map<String, String> tags = allParams.entrySet().stream()
                .filter(e -> !e.getKey().startsWith("_"))
                .collect(java.util.stream.Collectors.toMap(Map.Entry::getKey, Map.Entry::getValue));
        
        List<ServiceInstance> instances = tags.isEmpty()
                ? serviceRegistry.getAllInstances()
                : serviceRegistry.getAllInstancesWithTags(tags);
        return ResponseEntity.ok(instances);
    }

    @GetMapping("/services/{serviceName}/{instanceId}")
    public ResponseEntity<ServiceInstance> getInstance(
            @PathVariable String serviceName,
            @PathVariable String instanceId) {
        return serviceRegistry.getInstance(serviceName, instanceId)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/discover/{serviceName}")
    public ResponseEntity<List<ServiceInstance>> discover(@PathVariable String serviceName) {
        List<ServiceInstance> instances = serviceRegistry.getHealthyInstances(serviceName);
        return ResponseEntity.ok(instances);
    }

    @PostMapping("/subscribe")
    public ResponseEntity<Void> subscribe(@Valid @RequestBody SubscribeRequest request) {
        Subscriber subscriber = Subscriber.builder()
                .subscriberId(request.getSubscriberId())
                .serviceName(request.getServiceName())
                .callbackUrl(request.getCallbackUrl())
                .build();
        serviceRegistry.subscribe(subscriber);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/unsubscribe/{serviceName}/{subscriberId}")
    public ResponseEntity<Void> unsubscribe(
            @PathVariable String serviceName,
            @PathVariable String subscriberId) {
        serviceRegistry.unsubscribe(subscriberId, serviceName);
        return ResponseEntity.ok().build();
    }
}
