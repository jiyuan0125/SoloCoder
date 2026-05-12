package com.logaggregate.controller;

import com.logaggregate.model.RetentionConfig;
import com.logaggregate.service.retention.RetentionService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/logs/retention")
public class RetentionController {

    private final RetentionService retentionService;

    public RetentionController(RetentionService retentionService) {
        this.retentionService = retentionService;
    }

    @GetMapping
    public Map<String, RetentionConfig> getAllRetentionConfigs() {
        return retentionService.getAllRetentionConfigs();
    }

    @GetMapping("/{service}")
    public ResponseEntity<RetentionConfig> getRetentionConfig(@PathVariable String service) {
        RetentionConfig config = retentionService.getRetentionConfig(service);
        return ResponseEntity.ok(config);
    }

    @PutMapping("/{service}")
    public ResponseEntity<RetentionConfig> setRetentionDays(
            @PathVariable String service,
            @RequestBody Map<String, Integer> request
    ) {
        Integer days = request.get("retention_days");
        if (days == null) {
            days = request.get("retentionDays");
        }
        if (days == null) {
            return ResponseEntity.badRequest().build();
        }
        try {
            retentionService.setRetentionDays(service, days);
            return ResponseEntity.ok(new RetentionConfig(service, days));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().build();
        }
    }
}
