package com.loadbalancer.api;

import com.loadbalancer.dto.ApiResponse;
import com.loadbalancer.manager.SessionMigrationManager;
import com.loadbalancer.model.Node;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/route")
public class RouterController {

    private final SessionMigrationManager sessionManager;

    @Autowired
    public RouterController(SessionMigrationManager sessionManager) {
        this.sessionManager = sessionManager;
    }

    @GetMapping("/{sessionId}")
    public ResponseEntity<ApiResponse<Map<String, Object>>> routeSession(
            @PathVariable String sessionId) {
        Node target = sessionManager.route(sessionId);
        if (target == null) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                    .body(ApiResponse.error("No available nodes to route request"));
        }
        Map<String, Object> result = new HashMap<>();
        result.put("sessionId", sessionId);
        result.put("ip", target.getIp());
        result.put("port", target.getPort());
        result.put("nodeStatus", target.getStatus());
        return ResponseEntity.ok(ApiResponse.success(result));
    }

    @GetMapping("/stats")
    public ResponseEntity<ApiResponse<Map<String, Object>>> stats() {
        Map<String, Object> result = new HashMap<>();
        result.put("sessionMappingCount", sessionManager.getMappingCount());
        result.put("windowSeconds", sessionManager.getWindowSeconds());
        return ResponseEntity.ok(ApiResponse.success(result));
    }
}
