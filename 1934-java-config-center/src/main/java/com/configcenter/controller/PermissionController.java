package com.configcenter.controller;

import com.configcenter.dto.SetPermissionRequest;
import com.configcenter.exception.AppNotFoundException;
import com.configcenter.exception.ForbiddenException;
import com.configcenter.model.Permission;
import com.configcenter.model.Role;
import com.configcenter.service.AppService;
import com.configcenter.service.PermissionService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/apps/{appName}/permissions")
@RequiredArgsConstructor
public class PermissionController {

    private final AppService appService;
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

    private Role parseRole(String roleStr) {
        try {
            return Role.valueOf(roleStr.toUpperCase());
        } catch (IllegalArgumentException e) {
            throw new IllegalArgumentException("Invalid role: " + roleStr + ". Valid roles: READ_ONLY, READ_WRITE, ADMIN");
        }
    }

    @GetMapping
    public ResponseEntity<Map<String, Permission>> listPermissions(@PathVariable String appName,
                                                                    @RequestHeader(value = "X-User-Id", required = false) String userId) {
        String actualUserId = getUserId(userId);
        checkAppExists(appName);
        
        if (!permissionService.hasPermission(appName, actualUserId, Role.ADMIN)) {
            throw new ForbiddenException("No permission to manage permissions for app: " + appName);
        }
        
        return ResponseEntity.ok(permissionService.getAppPermissions(appName));
    }

    @PutMapping
    public ResponseEntity<Permission> setPermission(@PathVariable String appName,
                                                    @Valid @RequestBody SetPermissionRequest request,
                                                    @RequestHeader(value = "X-User-Id", required = false) String userId) {
        String actualUserId = getUserId(userId);
        checkAppExists(appName);
        
        if (!permissionService.hasPermission(appName, actualUserId, Role.ADMIN)) {
            throw new ForbiddenException("No permission to manage permissions for app: " + appName);
        }
        
        Role role = parseRole(request.getRole());
        permissionService.setPermission(appName, request.getUserId(), role, actualUserId);
        
        return permissionService.getPermission(appName, request.getUserId())
                .map(ResponseEntity::ok)
                .orElseThrow(() -> new RuntimeException("Failed to set permission"));
    }

    @DeleteMapping("/{targetUserId}")
    public ResponseEntity<Void> deletePermission(@PathVariable String appName,
                                                  @PathVariable String targetUserId,
                                                  @RequestHeader(value = "X-User-Id", required = false) String userId) {
        String actualUserId = getUserId(userId);
        checkAppExists(appName);
        
        if (!permissionService.hasPermission(appName, actualUserId, Role.ADMIN)) {
            throw new ForbiddenException("No permission to manage permissions for app: " + appName);
        }
        
        permissionService.deletePermission(appName, targetUserId, actualUserId);
        return ResponseEntity.noContent().build();
    }
}
