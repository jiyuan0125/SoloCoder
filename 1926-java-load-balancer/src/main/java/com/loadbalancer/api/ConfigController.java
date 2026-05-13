package com.loadbalancer.api;

import com.loadbalancer.dto.ApiResponse;
import com.loadbalancer.dto.ConfigRequest;
import com.loadbalancer.manager.SessionMigrationManager;
import com.loadbalancer.strategy.ConsistentHashRing;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/config")
public class ConfigController {

    private final ConsistentHashRing hashRing;
    private final SessionMigrationManager sessionManager;

    @Autowired
    public ConfigController(ConsistentHashRing hashRing, SessionMigrationManager sessionManager) {
        this.hashRing = hashRing;
        this.sessionManager = sessionManager;
    }

    @GetMapping
    public ResponseEntity<ApiResponse<Map<String, Object>>> getConfig() {
        Map<String, Object> config = new HashMap<>();
        config.put("vnodes_per_node", hashRing.getVnodesPerNode());
        config.put("session_window_seconds", sessionManager.getWindowSeconds());
        return ResponseEntity.ok(ApiResponse.success(config));
    }

    @PutMapping
    public ResponseEntity<ApiResponse<Map<String, Object>>> updateConfig(
            @Valid @RequestBody ConfigRequest request) {
        if (request.getVnodes_per_node() != null) {
            hashRing.setVnodesPerNode(request.getVnodes_per_node());
        }
        if (request.getSession_window_seconds() != null) {
            sessionManager.setWindowSeconds(request.getSession_window_seconds());
        }
        Map<String, Object> config = new HashMap<>();
        config.put("vnodes_per_node", hashRing.getVnodesPerNode());
        config.put("session_window_seconds", sessionManager.getWindowSeconds());
        return ResponseEntity.ok(ApiResponse.success("Config updated", config));
    }
}
