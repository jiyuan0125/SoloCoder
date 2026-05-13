package com.servicemesh.sidecar.controller;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.servicemesh.common.model.DiscoveryInstance;
import com.servicemesh.sidecar.service.DiscoveryService;
import com.servicemesh.sidecar.service.RegistrationService;
import com.servicemesh.sidecar.service.ReverseProxyService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.util.List;
import java.util.Map;

@RestController
public class SidecarController {

    private final RegistrationService registrationService;
    private final DiscoveryService discoveryService;
    private final ReverseProxyService reverseProxyService;
    private final ObjectMapper objectMapper = new ObjectMapper();

    public SidecarController(RegistrationService registrationService,
                             DiscoveryService discoveryService,
                             ReverseProxyService reverseProxyService) {
        this.registrationService = registrationService;
        this.discoveryService = discoveryService;
        this.reverseProxyService = reverseProxyService;
    }

    @GetMapping("/sidecar/health")
    public ResponseEntity<Map<String, Object>> health() {
        return ResponseEntity.ok(Map.of(
                "status", "UP",
                "registered", registrationService.isRegistered(),
                "instanceId", registrationService.getInstanceId()
        ));
    }

    @GetMapping("/sidecar/instances/{serviceName}")
    public ResponseEntity<List<DiscoveryInstance>> getInstances(@PathVariable String serviceName) {
        discoveryService.discoverService(serviceName);
        return ResponseEntity.ok(discoveryService.getInstances(serviceName));
    }

    @PostMapping("/sidecar/discovery/update")
    public ResponseEntity<Map<String, String>> updateDiscovery(@RequestBody String body) {
        try {
            JsonNode root = objectMapper.readTree(body);
            String serviceName = root.get("serviceName").asText();
            JsonNode instancesNode = root.get("instances");
            List<DiscoveryInstance> instances = new java.util.ArrayList<>();
            for (JsonNode node : instancesNode) {
                DiscoveryInstance inst = objectMapper.treeToValue(node, DiscoveryInstance.class);
                instances.add(inst);
            }
            discoveryService.updateServiceInstances(serviceName, instances);
            return ResponseEntity.ok(Map.of("status", "ok"));
        } catch (Exception e) {
            return ResponseEntity.status(400).body(Map.of("status", "error", "message", e.getMessage()));
        }
    }

    @PostMapping("/proxy/inbound/**")
    public void proxyInbound(HttpServletRequest request, HttpServletResponse response) throws Exception {
        reverseProxyService.proxyInbound(request, response);
    }

    @GetMapping("/proxy/inbound/**")
    public void proxyInboundGet(HttpServletRequest request, HttpServletResponse response) throws Exception {
        reverseProxyService.proxyInbound(request, response);
    }

    @PutMapping("/proxy/inbound/**")
    public void proxyInboundPut(HttpServletRequest request, HttpServletResponse response) throws Exception {
        reverseProxyService.proxyInbound(request, response);
    }

    @DeleteMapping("/proxy/inbound/**")
    public void proxyInboundDelete(HttpServletRequest request, HttpServletResponse response) throws Exception {
        reverseProxyService.proxyInbound(request, response);
    }

    @PostMapping("/proxy/outbound/**")
    public void proxyOutbound(HttpServletRequest request, HttpServletResponse response) throws Exception {
        String path = request.getRequestURI().replace("/proxy/outbound", "");
        reverseProxyService.proxyOutboundWithPath(request, response, path);
    }

    @GetMapping("/proxy/outbound/**")
    public void proxyOutboundGet(HttpServletRequest request, HttpServletResponse response) throws Exception {
        String path = request.getRequestURI().replace("/proxy/outbound", "");
        reverseProxyService.proxyOutboundWithPath(request, response, path);
    }

    @PutMapping("/proxy/outbound/**")
    public void proxyOutboundPut(HttpServletRequest request, HttpServletResponse response) throws Exception {
        String path = request.getRequestURI().replace("/proxy/outbound", "");
        reverseProxyService.proxyOutboundWithPath(request, response, path);
    }

    @DeleteMapping("/proxy/outbound/**")
    public void proxyOutboundDelete(HttpServletRequest request, HttpServletResponse response) throws Exception {
        String path = request.getRequestURI().replace("/proxy/outbound", "");
        reverseProxyService.proxyOutboundWithPath(request, response, path);
    }
}
