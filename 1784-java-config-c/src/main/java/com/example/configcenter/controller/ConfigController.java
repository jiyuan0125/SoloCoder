package com.example.configcenter.controller;

import com.example.configcenter.dto.BatchImportResponse;
import com.example.configcenter.dto.ConfigRequest;
import com.example.configcenter.dto.ConfigResponse;
import com.example.configcenter.dto.ErrorResponse;
import com.example.configcenter.dto.SubscriberRequest;
import com.example.configcenter.entity.ConfigHistory;
import com.example.configcenter.service.ConfigService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/configs")
@RequiredArgsConstructor
public class ConfigController {
    
    private final ConfigService configService;
    
    @PostMapping
    public ResponseEntity<ConfigResponse> createConfig(@RequestBody ConfigRequest request) {
        ConfigResponse response = configService.saveConfig(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }
    
    @PutMapping("/{key}")
    public ResponseEntity<ConfigResponse> updateConfig(
            @PathVariable String key, 
            @RequestBody ConfigRequest request) {
        request.setKey(key);
        ConfigResponse response = configService.saveConfig(request);
        return ResponseEntity.ok(response);
    }
    
    @GetMapping("/{key}")
    public ResponseEntity<ConfigResponse> getConfig(@PathVariable String key) {
        return configService.getConfig(key)
            .map(ResponseEntity::ok)
            .orElse(ResponseEntity.notFound().build());
    }
    
    @DeleteMapping("/{key}")
    public ResponseEntity<Void> deleteConfig(@PathVariable String key) {
        boolean deleted = configService.deleteConfig(key);
        if (deleted) {
            return ResponseEntity.noContent().build();
        }
        return ResponseEntity.notFound().build();
    }
    
    @PostMapping("/batch")
    public ResponseEntity<BatchImportResponse> batchImport(@RequestBody List<ConfigRequest> requests) {
        BatchImportResponse response = configService.batchImport(requests);
        return ResponseEntity.ok(response);
    }
    
    @GetMapping("/{key}/history")
    public ResponseEntity<List<ConfigHistory>> getHistory(
            @PathVariable String key,
            @RequestParam(required = false) String startTime,
            @RequestParam(required = false) String endTime) {
        List<ConfigHistory> history = configService.getHistory(key, startTime, endTime);
        return ResponseEntity.ok(history);
    }
    
    @PostMapping("/subscribe")
    public ResponseEntity<Void> subscribe(@RequestBody SubscriberRequest request) {
        configService.subscribe(request.getKey(), request.getCallbackUrl());
        return ResponseEntity.ok().build();
    }
}
