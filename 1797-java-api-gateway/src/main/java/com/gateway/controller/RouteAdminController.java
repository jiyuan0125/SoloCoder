package com.gateway.controller;

import com.gateway.model.RouteRule;
import com.gateway.router.RouteStore;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("${gateway.admin.prefix:/admin}/routes")
public class RouteAdminController {

    private final RouteStore routeStore;

    public RouteAdminController(RouteStore routeStore) {
        this.routeStore = routeStore;
    }

    @GetMapping
    public ResponseEntity<List<RouteRule>> listRoutes() {
        return ResponseEntity.ok(routeStore.getAllRoutes());
    }

    @GetMapping("/{id}")
    public ResponseEntity<RouteRule> getRoute(@PathVariable String id) {
        RouteRule route = routeStore.getRoute(id);
        if (route == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(route);
    }

    @PostMapping
    public ResponseEntity<Map<String, Object>> createRoute(@Valid @RequestBody RouteRule route) {
        routeStore.addRoute(route);
        
        Map<String, Object> result = new HashMap<>();
        result.put("message", "路由规则已创建");
        result.put("routeId", route.getId());
        
        return ResponseEntity.ok(result);
    }

    @PutMapping("/{id}")
    public ResponseEntity<Map<String, Object>> updateRoute(
            @PathVariable String id,
            @Valid @RequestBody RouteRule route) {
        if (!id.equals(route.getId())) {
            return ResponseEntity.badRequest().build();
        }
        
        routeStore.addRoute(route);
        
        Map<String, Object> result = new HashMap<>();
        result.put("message", "路由规则已更新");
        result.put("routeId", route.getId());
        
        return ResponseEntity.ok(result);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Map<String, Object>> deleteRoute(@PathVariable String id) {
        boolean deleted = routeStore.deleteRoute(id);
        
        Map<String, Object> result = new HashMap<>();
        if (deleted) {
            result.put("message", "路由规则已删除");
            return ResponseEntity.ok(result);
        } else {
            result.put("message", "路由规则不存在");
            return ResponseEntity.notFound().build();
        }
    }
}
