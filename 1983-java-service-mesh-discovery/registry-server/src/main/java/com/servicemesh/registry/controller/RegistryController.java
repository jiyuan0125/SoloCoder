package com.servicemesh.registry.controller;

import com.servicemesh.common.model.RegisterRequest;
import com.servicemesh.common.model.ServiceInstance;
import com.servicemesh.registry.service.RegistryService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/registry")
public class RegistryController {

    private final RegistryService registryService;

    public RegistryController(RegistryService registryService) {
        this.registryService = registryService;
    }

    @PostMapping("/register")
    public ResponseEntity<Map<String, String>> register(@RequestBody RegisterRequest request) {
        ServiceInstance instance = new ServiceInstance(
                request.getServiceName(),
                null,
                request.getIp(),
                request.getPort(),
                request.getVersion(),
                request.getZone()
        );
        String instanceId = registryService.register(instance);
        return ResponseEntity.ok(Map.of("instanceId", instanceId, "status", "ok"));
    }

    @PostMapping("/heartbeat")
    public ResponseEntity<Map<String, String>> heartbeat(@RequestBody Map<String, String> request) {
        String serviceName = request.get("serviceName");
        String instanceId = request.get("instanceId");
        boolean ok = registryService.heartbeat(serviceName, instanceId);
        return ok ? ResponseEntity.ok(Map.of("status", "ok"))
                  : ResponseEntity.status(404).body(Map.of("status", "not_found"));
    }

    @PostMapping("/deregister")
    public ResponseEntity<Map<String, String>> deregister(@RequestBody Map<String, String> request) {
        String serviceName = request.get("serviceName");
        String instanceId = request.get("instanceId");
        registryService.deregister(serviceName, instanceId);
        return ResponseEntity.ok(Map.of("status", "ok"));
    }

    @GetMapping("/instances/{serviceName}")
    public ResponseEntity<List<ServiceInstance>> getInstances(@PathVariable String serviceName,
                                                              @RequestParam(defaultValue = "true") boolean healthy) {
        return ResponseEntity.ok(registryService.getInstances(serviceName, healthy));
    }

    @GetMapping("/instances")
    public ResponseEntity<List<ServiceInstance>> getAllInstances(@RequestParam(defaultValue = "false") boolean healthy) {
        return ResponseEntity.ok(registryService.getAllInstances(healthy));
    }

    @GetMapping("/services")
    public ResponseEntity<List<String>> getServices() {
        return ResponseEntity.ok(registryService.getAllServiceNames());
    }

    @PostMapping("/subscribe")
    public ResponseEntity<Map<String, String>> subscribe(@RequestBody Map<String, String> request) {
        registryService.subscribe(request.get("serviceName"), request.get("callbackUrl"));
        return ResponseEntity.ok(Map.of("status", "ok"));
    }

    @PostMapping("/unsubscribe")
    public ResponseEntity<Map<String, String>> unsubscribe(@RequestBody Map<String, String> request) {
        registryService.unsubscribe(request.get("serviceName"), request.get("callbackUrl"));
        return ResponseEntity.ok(Map.of("status", "ok"));
    }
}
