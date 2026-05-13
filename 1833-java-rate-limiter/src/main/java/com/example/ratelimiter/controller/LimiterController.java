package com.example.ratelimiter.controller;

import com.example.ratelimiter.core.BucketManager;
import com.example.ratelimiter.model.BucketStats;
import com.example.ratelimiter.model.LimiterConfig;
import com.example.ratelimiter.model.UpdateLimiterConfigRequest;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.validation.Valid;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/limiter")
public class LimiterController {

    private static final String CONFIG_PREFIX = "/limiter/config";

    private final BucketManager bucketManager;

    public LimiterController(BucketManager bucketManager) {
        this.bucketManager = bucketManager;
    }

    @GetMapping("/stats")
    public ResponseEntity<List<BucketStats>> getStats() {
        return ResponseEntity.ok(bucketManager.getAllStats());
    }

    @PutMapping("/config")
    public ResponseEntity<Void> updateConfigs(@RequestBody @Valid List<LimiterConfig> configs) {
        bucketManager.updateConfigs(configs);
        return ResponseEntity.ok().build();
    }

    @PatchMapping("/config/**")
    public ResponseEntity<Void> updateConfig(
            HttpServletRequest request,
            @RequestBody @Valid UpdateLimiterConfigRequest updateRequest) {
        
        String path = extractPath(request);
        if (path == null || path.isEmpty()) {
            return ResponseEntity.badRequest().build();
        }
        LimiterConfig config = new LimiterConfig(path, updateRequest.getCapacity(), updateRequest.getRate());
        bucketManager.updateConfig(config);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/config/**")
    public ResponseEntity<Void> deleteConfig(HttpServletRequest request) {
        String path = extractPath(request);
        if (path == null || path.isEmpty()) {
            return ResponseEntity.badRequest().build();
        }
        boolean deleted = bucketManager.deleteConfig(path);
        if (deleted) {
            return ResponseEntity.ok().build();
        }
        return ResponseEntity.notFound().build();
    }

    private String extractPath(HttpServletRequest request) {
        String uri = request.getRequestURI();
        String contextPath = request.getContextPath();
        if (contextPath != null && !contextPath.isEmpty() && uri.startsWith(contextPath)) {
            uri = uri.substring(contextPath.length());
        }
        if (uri.startsWith(CONFIG_PREFIX)) {
            return uri.substring(CONFIG_PREFIX.length());
        }
        return null;
    }
}
