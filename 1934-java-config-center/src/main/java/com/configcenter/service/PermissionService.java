package com.configcenter.service;

import com.configcenter.model.App;
import com.configcenter.model.AuditAction;
import com.configcenter.model.AuditLog;
import com.configcenter.model.Permission;
import com.configcenter.model.Role;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.Map;
import java.util.Optional;

@Service
@RequiredArgsConstructor
public class PermissionService {

    private final DataStore dataStore;
    private final AuditService auditService;

    public void setPermission(String appName, String targetUserId, Role role, String currentUserId) {
        Map<String, Permission> appPermissions = dataStore.getPermissions()
                .computeIfAbsent(appName, k -> new java.util.concurrent.ConcurrentHashMap<>());
        
        String oldRole = null;
        Permission existing = appPermissions.get(targetUserId);
        if (existing != null) {
            oldRole = existing.getRole().name();
        }
        
        Permission permission = new Permission(appName, targetUserId, role);
        appPermissions.put(targetUserId, permission);
        
        String newValue = role.name() + " for user: " + targetUserId;
        String oldValue = oldRole != null ? oldRole + " for user: " + targetUserId : null;
        
        AuditLog log = auditService.createLog(currentUserId, appName, AuditAction.SET_PERMISSION,
                oldValue, newValue, null);
        auditService.saveLog(log);
    }

    public Optional<Permission> getPermission(String appName, String userId) {
        Map<String, Permission> appPermissions = dataStore.getPermissions().get(appName);
        if (appPermissions == null) {
            return Optional.empty();
        }
        return Optional.ofNullable(appPermissions.get(userId));
    }

    public Map<String, Permission> getAppPermissions(String appName) {
        Map<String, Permission> appPermissions = dataStore.getPermissions().get(appName);
        if (appPermissions == null) {
            return Map.of();
        }
        return appPermissions;
    }

    public void deletePermission(String appName, String targetUserId, String currentUserId) {
        Map<String, Permission> appPermissions = dataStore.getPermissions().get(appName);
        if (appPermissions != null) {
            Permission removed = appPermissions.remove(targetUserId);
            if (removed != null) {
                AuditLog log = auditService.createLog(currentUserId, appName, AuditAction.DELETE_PERMISSION,
                        removed.getRole().name() + " for user: " + targetUserId, null, null);
                auditService.saveLog(log);
            }
        }
    }

    public boolean hasPermission(String appName, String userId, Role requiredRole) {
        Optional<Permission> permission = getPermission(appName, userId);
        if (permission.isEmpty()) {
            return false;
        }
        
        Role userRole = permission.get().getRole();
        
        return switch (requiredRole) {
            case READ_ONLY -> true;
            case READ_WRITE -> userRole == Role.READ_WRITE || userRole == Role.ADMIN;
            case ADMIN -> userRole == Role.ADMIN;
        };
    }
}
