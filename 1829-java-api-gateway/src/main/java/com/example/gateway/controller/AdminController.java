package com.example.gateway.controller;

import com.example.gateway.model.BackendTarget;
import com.example.gateway.model.RouteRule;
import com.example.gateway.model.TestResult;
import com.example.gateway.service.RouteService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.*;

@RestController
@RequestMapping("/admin")
public class AdminController {

    private final RouteService routeService;

    public AdminController(RouteService routeService) {
        this.routeService = routeService;
    }

    @GetMapping("/routes")
    public ResponseEntity<Collection<RouteRule>> getAllRoutes() {
        return ResponseEntity.ok(routeService.getAllRoutes());
    }

    @GetMapping("/routes/{id}")
    public ResponseEntity<RouteRule> getRoute(@PathVariable String id) {
        RouteRule route = routeService.getRoute(id);
        if (route == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(route);
    }

    @PostMapping("/routes")
    public ResponseEntity<RouteRule> createRoute(@RequestBody RouteRule route) {
        if (route.getId() == null || route.getId().isEmpty()) {
            route.setId(UUID.randomUUID().toString());
        }
        routeService.addRoute(route);
        return ResponseEntity.status(HttpStatus.CREATED).body(route);
    }

    @PutMapping("/routes/{id}")
    public ResponseEntity<RouteRule> updateRoute(@PathVariable String id, @RequestBody RouteRule route) {
        boolean updated = routeService.updateRoute(id, route);
        if (!updated) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(routeService.getRoute(id));
    }

    @DeleteMapping("/routes/{id}")
    public ResponseEntity<Void> deleteRoute(@PathVariable String id) {
        boolean deleted = routeService.deleteRoute(id);
        if (!deleted) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.noContent().build();
    }

    @PutMapping("/routes/{id}/filters")
    public ResponseEntity<RouteRule> updateFilterChain(
            @PathVariable String id,
            @RequestBody List<String> filterNames) {
        RouteRule existing = routeService.getRoute(id);
        if (existing == null) {
            return ResponseEntity.notFound().build();
        }
        routeService.updateFilterChain(id, filterNames);
        return ResponseEntity.ok(routeService.getRoute(id));
    }

    @GetMapping("/routes/{id}/filters")
    public ResponseEntity<List<String>> getFilterChain(@PathVariable String id) {
        RouteRule route = routeService.getRoute(id);
        if (route == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(route.getFilterNames() != null ? route.getFilterNames() : new ArrayList<>());
    }

    @GetMapping("/test")
    public ResponseEntity<TestResult> testRoute(@RequestParam String path) {
        RouteRule route = routeService.matchRoute(path);
        if (route == null) {
            return ResponseEntity.ok(new TestResult(null, null, null, null));
        }
        
        BackendTarget backend = routeService.selectBackend(route);
        String backendUrl = backend != null ? backend.getUrl() : null;
        
        TestResult result = new TestResult(
            route.getId(),
            route.getPathPrefix(),
            backendUrl,
            route.getFilterNames() != null ? route.getFilterNames() : new ArrayList<>()
        );
        return ResponseEntity.ok(result);
    }
}
