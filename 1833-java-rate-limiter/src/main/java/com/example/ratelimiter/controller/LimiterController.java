package com.example.ratelimiter.controller;

import com.example.ratelimiter.core.BucketManager;
import com.example.ratelimiter.model.BucketStats;
import com.example.ratelimiter.model.LimiterConfig;
import jakarta.validation.Valid;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/limiter")
public class LimiterController {

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

    @PatchMapping("/config/{path}")
    public ResponseEntity<Void> updateConfig(
            @PathVariable("path") String path,
            @RequestBody @Valid LimiterConfig config) {
        
        if (!path.equals(config.getPath())) {
            config.setPath(path);
        }
        bucketManager.updateConfig(config);
        return ResponseEntity.ok().build();
    }

    @DeleteMapping("/config/{path}")
    public ResponseEntity<Void> deleteConfig(@PathVariable("path") String path) {
        boolean deleted = bucketManager.deleteConfig(path);
        if (deleted) {
            return ResponseEntity.ok().build();
        }
        return ResponseEntity.notFound().build();
    }
}
