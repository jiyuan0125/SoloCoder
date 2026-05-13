package com.configcenter.controller;

import com.configcenter.config.LongPollConfig;
import com.configcenter.dto.CreateConfigRequest;
import com.configcenter.dto.RollbackConfigRequest;
import com.configcenter.dto.UpdateConfigRequest;
import com.configcenter.exception.InvalidRequestException;
import com.configcenter.model.AuditLog;
import com.configcenter.model.ConfigHistory;
import com.configcenter.model.ConfigItem;
import com.configcenter.service.ConfigChangeNotifier;
import com.configcenter.service.ConfigService;
import com.configcenter.service.LongPollSubscriber;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDateTime;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.TimeUnit;

import org.springframework.format.annotation.DateTimeFormat;

@RestController
@RequestMapping("/api/configs")
@RequiredArgsConstructor
public class ConfigController {
    
    private final ConfigService configService;
    private final ConfigChangeNotifier configChangeNotifier;
    private final LongPollConfig longPollConfig;
    
    @GetMapping
    public ResponseEntity<Page<ConfigItem>> listConfigs(
            @RequestParam(required = false) String namespace,
            @RequestParam(required = false) String group,
            @RequestParam(required = false) String key,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size) {
        Pageable pageable = PageRequest.of(page, size);
        Page<ConfigItem> configs = configService.listConfigs(namespace, group, key, pageable);
        return ResponseEntity.ok(configs);
    }
    
    @GetMapping("/{namespace}/{group}/{key}")
    public ResponseEntity<ConfigItem> getConfig(
            @PathVariable String namespace,
            @PathVariable String group,
            @PathVariable String key) {
        return configService.getConfig(namespace, group, key)
                .map(ResponseEntity::ok)
                .orElseGet(() -> ResponseEntity.notFound().build());
    }
    
    @PostMapping
    public ResponseEntity<ConfigItem> createConfig(
            @Valid @RequestBody CreateConfigRequest request,
            HttpServletRequest httpRequest) {
        String clientIp = getClientIp(httpRequest);
        String operator = getOperator(request.getOperator());
        ConfigItem created = configService.createConfig(
                request.getNamespace(),
                request.getGroup(),
                request.getKey(),
                request.getValue(),
                operator,
                clientIp
        );
        return ResponseEntity.status(HttpStatus.CREATED).body(created);
    }
    
    @PutMapping("/{namespace}/{group}/{key}")
    public ResponseEntity<ConfigItem> updateConfig(
            @PathVariable String namespace,
            @PathVariable String group,
            @PathVariable String key,
            @RequestBody UpdateConfigRequest request,
            HttpServletRequest httpRequest) {
        String clientIp = getClientIp(httpRequest);
        String operator = getOperator(request.getOperator());
        ConfigItem updated = configService.updateConfig(
                namespace, group, key,
                request.getValue(),
                operator, clientIp
        );
        return ResponseEntity.ok(updated);
    }
    
    @DeleteMapping("/{namespace}/{group}/{key}")
    public ResponseEntity<Void> deleteConfig(
            @PathVariable String namespace,
            @PathVariable String group,
            @PathVariable String key,
            @RequestParam(required = false) String operator,
            HttpServletRequest httpRequest) {
        String clientIp = getClientIp(httpRequest);
        String op = getOperator(operator);
        configService.deleteConfig(namespace, group, key, op, clientIp);
        return ResponseEntity.noContent().build();
    }
    
    @PostMapping("/{namespace}/{group}/{key}/rollback")
    public ResponseEntity<ConfigItem> rollbackConfig(
            @PathVariable String namespace,
            @PathVariable String group,
            @PathVariable String key,
            @RequestBody RollbackConfigRequest request,
            HttpServletRequest httpRequest) {
        if (request.getTargetVersion() == null) {
            throw new InvalidRequestException("targetVersion is required");
        }
        String clientIp = getClientIp(httpRequest);
        String operator = getOperator(request.getOperator());
        ConfigItem rolledBack = configService.rollbackConfig(
                namespace, group, key,
                request.getTargetVersion(),
                operator, clientIp
        );
        return ResponseEntity.ok(rolledBack);
    }
    
    @GetMapping("/{namespace}/{group}/{key}/history")
    public ResponseEntity<Page<ConfigHistory>> listHistory(
            @PathVariable String namespace,
            @PathVariable String group,
            @PathVariable String key,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime startTime,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime endTime,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size) {
        Pageable pageable = PageRequest.of(page, size);
        Page<ConfigHistory> history = configService.listHistory(namespace, group, key, startTime, endTime, pageable);
        return ResponseEntity.ok(history);
    }
    
    @GetMapping("/{namespace}/{group}/{key}/history/{version}")
    public ResponseEntity<ConfigHistory> getHistoryVersion(
            @PathVariable String namespace,
            @PathVariable String group,
            @PathVariable String key,
            @PathVariable Long version) {
        return configService.getHistoryVersion(namespace, group, key, version)
                .map(ResponseEntity::ok)
                .orElseGet(() -> ResponseEntity.notFound().build());
    }
    
    @GetMapping("/{namespace}/{group}/{key}/audit")
    public ResponseEntity<Page<AuditLog>> listAuditLogs(
            @PathVariable String namespace,
            @PathVariable String group,
            @PathVariable String key,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime startTime,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE_TIME) LocalDateTime endTime,
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "20") int size) {
        Pageable pageable = PageRequest.of(page, size);
        Page<AuditLog> logs = configService.listAuditLogs(namespace, group, key, startTime, endTime, pageable);
        return ResponseEntity.ok(logs);
    }
    
    @GetMapping("/{namespace}/{group}/{key}/watch")
    public ResponseEntity<Map<String, Object>> watchConfig(
            @PathVariable String namespace,
            @PathVariable String group,
            @PathVariable String key,
            @RequestParam(required = false) Long lastVersion,
            HttpServletRequest httpRequest) {
        configService.validateKeyParams(namespace, group, key);
        
        Optional<ConfigItem> currentOpt = configService.getConfigFresh(namespace, group, key);
        
        if (currentOpt.isPresent() && lastVersion != null) {
            ConfigItem current = currentOpt.get();
            if (current.getVersion() > lastVersion) {
                Map<String, Object> result = buildChangeResponse(current);
                return ResponseEntity.ok(result);
            }
        }
        
        LongPollSubscriber subscriber = configChangeNotifier.subscribe(namespace, group, key);
        try {
            boolean changed = subscriber.waitForChange(longPollConfig.getTimeoutSeconds(), TimeUnit.SECONDS);
            
            if (changed) {
                Optional<ConfigItem> updated = configService.getConfigFresh(namespace, group, key);
                if (updated.isPresent()) {
                    Map<String, Object> result = buildChangeResponse(updated.get());
                    return ResponseEntity.ok(result);
                } else {
                    Map<String, Object> result = new HashMap<>();
                    result.put("deleted", true);
                    result.put("namespace", namespace);
                    result.put("group", group);
                    result.put("key", key);
                    return ResponseEntity.ok(result);
                }
            } else {
                return ResponseEntity.status(HttpStatus.NOT_MODIFIED).build();
            }
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        } finally {
            configChangeNotifier.unsubscribe(namespace, group, key, subscriber);
        }
    }
    
    private Map<String, Object> buildChangeResponse(ConfigItem config) {
        Map<String, Object> result = new HashMap<>();
        result.put("namespace", config.getNamespace());
        result.put("group", config.getGroup());
        result.put("key", config.getKey());
        result.put("value", config.getValue());
        result.put("version", config.getVersion());
        result.put("updatedAt", config.getUpdatedAt());
        return result;
    }
    
    private String getClientIp(HttpServletRequest request) {
        String ip = request.getHeader("X-Forwarded-For");
        if (ip == null || ip.isEmpty() || "unknown".equalsIgnoreCase(ip)) {
            ip = request.getHeader("Proxy-Client-IP");
        }
        if (ip == null || ip.isEmpty() || "unknown".equalsIgnoreCase(ip)) {
            ip = request.getHeader("WL-Proxy-Client-IP");
        }
        if (ip == null || ip.isEmpty() || "unknown".equalsIgnoreCase(ip)) {
            ip = request.getRemoteAddr();
        }
        if (ip != null && ip.contains(",")) {
            ip = ip.split(",")[0].trim();
        }
        return ip != null ? ip : "unknown";
    }
    
    private String getOperator(String operator) {
        return operator != null && !operator.trim().isEmpty() ? operator : "anonymous";
    }
}
