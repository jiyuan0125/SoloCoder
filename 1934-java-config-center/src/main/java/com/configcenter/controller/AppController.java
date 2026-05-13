package com.configcenter.controller;

import com.configcenter.dto.CreateAppRequest;
import com.configcenter.exception.AppAlreadyExistsException;
import com.configcenter.exception.AppNotFoundException;
import com.configcenter.exception.ForbiddenException;
import com.configcenter.model.App;
import com.configcenter.model.Role;
import com.configcenter.service.AppService;
import com.configcenter.service.PermissionService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/apps")
@RequiredArgsConstructor
public class AppController {

    private final AppService appService;
    private final PermissionService permissionService;

    private String getUserId(String userId) {
        if (userId == null || userId.isEmpty()) {
            throw new IllegalArgumentException("X-User-Id header is required");
        }
        return userId;
    }

    @PostMapping
    public ResponseEntity<App> createApp(@Valid @RequestBody CreateAppRequest request,
                                         @RequestHeader(value = "X-User-Id", required = false) String userId) {
        String actualUserId = getUserId(userId);
        
        if (appService.appExists(request.getName())) {
            throw new AppAlreadyExistsException("App already exists: " + request.getName());
        }
        
        App app = appService.createApp(request.getName(), request.getDescription(), actualUserId);
        
        permissionService.setPermission(request.getName(), actualUserId, Role.ADMIN, actualUserId);
        
        return ResponseEntity.ok(app);
    }

    @GetMapping("/{name}")
    public ResponseEntity<App> getApp(@PathVariable String name,
                                      @RequestHeader(value = "X-User-Id", required = false) String userId) {
        String actualUserId = getUserId(userId);
        
        if (!permissionService.hasPermission(name, actualUserId, Role.READ_ONLY)) {
            throw new ForbiddenException("No permission to access app: " + name);
        }
        
        return appService.getApp(name)
                .map(ResponseEntity::ok)
                .orElseThrow(() -> new AppNotFoundException("App not found: " + name));
    }

    @DeleteMapping("/{name}")
    public ResponseEntity<Void> deleteApp(@PathVariable String name,
                                          @RequestHeader(value = "X-User-Id", required = false) String userId) {
        String actualUserId = getUserId(userId);
        
        if (!appService.appExists(name)) {
            throw new AppNotFoundException("App not found: " + name);
        }
        
        if (!permissionService.hasPermission(name, actualUserId, Role.ADMIN)) {
            throw new ForbiddenException("No permission to delete app: " + name);
        }
        
        appService.deleteApp(name, actualUserId);
        return ResponseEntity.noContent().build();
    }

    @GetMapping
    public ResponseEntity<Map<String, App>> listApps(@RequestHeader(value = "X-User-Id", required = false) String userId) {
        String actualUserId = getUserId(userId);
        
        Map<String, App> accessibleApps = new HashMap<>();
        for (Map.Entry<String, App> entry : appService.getAllApps().entrySet()) {
            if (permissionService.hasPermission(entry.getKey(), actualUserId, Role.READ_ONLY)) {
                accessibleApps.put(entry.getKey(), entry.getValue());
            }
        }
        
        return ResponseEntity.ok(accessibleApps);
    }
}
