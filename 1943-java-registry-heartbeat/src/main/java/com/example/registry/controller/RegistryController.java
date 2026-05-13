package com.example.registry.controller;

import com.example.registry.dto.InstanceResponse;
import com.example.registry.dto.RegisterRequest;
import com.example.registry.dto.ServiceSummary;
import com.example.registry.dto.SubscribeRequest;
import com.example.registry.model.ServiceInstance;
import com.example.registry.service.InstanceService;
import com.example.registry.service.SubscriptionService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.stream.Collectors;

@RestController
@RequiredArgsConstructor
public class RegistryController {
    private final InstanceService instanceService;
    private final SubscriptionService subscriptionService;

    @PostMapping("/services/{name}/instances")
    public ResponseEntity<ServiceInstance> register(
            @PathVariable("name") String serviceName,
            @Valid @RequestBody RegisterRequest request) {
        ServiceInstance instance = instanceService.register(serviceName, request);
        return ResponseEntity.ok(instance);
    }

    @PutMapping("/instances/{id}/heartbeat")
    public ResponseEntity<Void> heartbeat(@PathVariable("id") String instanceId) {
        if (instanceService.heartbeat(instanceId).isPresent()) {
            return ResponseEntity.ok().build();
        }
        return ResponseEntity.notFound().build();
    }

    @GetMapping("/services/{name}/instances")
    public ResponseEntity<List<InstanceResponse>> getInstances(
            @PathVariable("name") String serviceName,
            @RequestParam(value = "healthy_only", required = false, defaultValue = "false") boolean healthyOnly) {
        List<ServiceInstance> instances = instanceService.getByServiceName(serviceName, healthyOnly);
        List<InstanceResponse> response = instances.stream()
                .map(InstanceResponse::from)
                .collect(Collectors.toList());
        return ResponseEntity.ok(response);
    }

    @GetMapping("/services")
    public ResponseEntity<List<ServiceSummary>> listServices() {
        List<ServiceSummary> services = instanceService.getAllServiceNames().stream()
                .map(name -> ServiceSummary.builder()
                        .serviceName(name)
                        .instanceCount(instanceService.countByServiceName(name))
                        .build())
                .collect(Collectors.toList());
        return ResponseEntity.ok(services);
    }

    @DeleteMapping("/instances/{id}")
    public ResponseEntity<Void> deregister(@PathVariable("id") String instanceId) {
        if (instanceService.findById(instanceId).isEmpty()) {
            return ResponseEntity.notFound().build();
        }
        instanceService.deregister(instanceId);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/subscriptions")
    public ResponseEntity<?> subscribe(@Valid @RequestBody SubscribeRequest request) {
        return ResponseEntity.ok(subscriptionService.subscribe(request));
    }
}
