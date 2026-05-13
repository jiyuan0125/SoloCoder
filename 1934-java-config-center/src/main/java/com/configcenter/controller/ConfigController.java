package com.configcenter.controller;

import com.configcenter.dto.ConfigResponse;
import com.configcenter.dto.CreateOrUpdateConfigRequest;
import com.configcenter.exception.AppNotFoundException;
import com.configcenter.exception.ConfigNotFoundException;
import com.configcenter.exception.ForbiddenException;
import com.configcenter.model.Config;
import com.configcenter.model.Role;
import com.configcenter.service.AppService;
import com.configcenter.service.ConfigService;
import com.configcenter.service.PermissionService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/apps/{appName}/configs")
@RequiredArgsConstructor
public class ConfigController {

    private final AppService appService;
    private final ConfigService configService;
    private final PermissionService permissionService;

    private String getUserId(String userId) {
        if (userId == null || userId.isEmpty()) {
            throw new IllegalArgumentException("X-User-Id header is required");
        }
        return userId;
    }

    private void checkAppExists(String appName) {
        if (!appService.appExists(appName)) {
            throw new AppNotFoundException("App not found: " + appName);
        }
    }

    @GetMapping
    public ResponseEntity<Map<String, ConfigResponse>> listConfigs(@PathVariable String appName,
                                                                   @RequestHeader(value = "X-User-Id", required = false) String userId) {
        String actualUserId = getUserId(userId);
        checkAppExists(appName);
        
        if (!permissionService.hasPermission(appName, actualUserId, Role.READ_ONLY)) {
            throw new ForbiddenException("No permission to read configs for app: " + appName);
        }
        
        Map<String, ConfigResponse> response = configService.getAllConfigs(appName).entrySet().stream()
                .collect(Collectors.toMap(
                        Map.Entry::getKey,
                        e -> new ConfigResponse(e.getValue().getKey(), e.getValue().getValue(), e.getValue().isSecret())
                ));
        
        return ResponseEntity.ok(response);
    }

    @GetMapping("/{key}")
    public ResponseEntity<ConfigResponse> getConfig(@PathVariable String appName,
                                                    @PathVariable String key,
                                                    @RequestHeader(value = "X-User-Id", required = false) String userId) {
        String actualUserId = getUserId(userId);
        checkAppExists(appName);
        
        if (!permissionService.hasPermission(appName, actualUserId, Role.READ_ONLY)) {
            throw new ForbiddenException("No permission to read configs for app: " + appName);
        }
        
        return configService.getConfig(appName, key)
                .map(c -> ResponseEntity.ok(new ConfigResponse(c.getKey(), c.getValue(), c.isSecret())))
                .orElseThrow(() -> new ConfigNotFoundException("Config not found: " + key));
    }

    @PostMapping
    public ResponseEntity<ConfigResponse> createConfig(@PathVariable String appName,
                                                       @Valid @RequestBody CreateOrUpdateConfigRequest request,
                                                       @RequestHeader(value = "X-User-Id", required = false) String userId) {
        String actualUserId = getUserId(userId);
        checkAppExists(appName);
        
        if (!permissionService.hasPermission(appName, actualUserId, Role.READ_WRITE)) {
            throw new ForbiddenException("No permission to modify configs for app: " + appName);
        }
        
        if (request.getKey() == null || request.getKey().isBlank()) {
            throw new IllegalArgumentException("Config key is required");
        }
        
        Config config = configService.createOrUpdateConfig(appName, request.getKey(), request.getValue(), request.isSecret(), actualUserId);
        return ResponseEntity.ok(new ConfigResponse(config.getKey(), config.getValue(), config.isSecret()));
    }

    @PutMapping("/{key}")
    public ResponseEntity<ConfigResponse> updateConfig(@PathVariable String appName,
                                                       @PathVariable String key,
                                                       @Valid @RequestBody CreateOrUpdateConfigRequest request,
                                                       @RequestHeader(value = "X-User-Id", required = false) String userId) {
        String actualUserId = getUserId(userId);
        checkAppExists(appName);
        
        if (!permissionService.hasPermission(appName, actualUserId, Role.READ_WRITE)) {
            throw new ForbiddenException("No permission to modify configs for app: " + appName);
        }
        
        Config config = configService.createOrUpdateConfig(appName, key, request.getValue(), request.isSecret(), actualUserId);
        return ResponseEntity.ok(new ConfigResponse(config.getKey(), config.getValue(), config.isSecret()));
    }

    @DeleteMapping("/{key}")
    public ResponseEntity<Void> deleteConfig(@PathVariable String appName,
                                             @PathVariable String key,
                                             @RequestHeader(value = "X-User-Id", required = false) String userId) {
        String actualUserId = getUserId(userId);
        checkAppExists(appName);
        
        if (!permissionService.hasPermission(appName, actualUserId, Role.READ_WRITE)) {
            throw new ForbiddenException("No permission to delete configs for app: " + appName);
        }
        
        if (!configService.configExists(appName, key)) {
            throw new ConfigNotFoundException("Config not found: " + key);
        }
        
        configService.deleteConfig(appName, key, actualUserId);
        return ResponseEntity.noContent().build();
    }
}
