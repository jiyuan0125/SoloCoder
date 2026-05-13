package com.configcenter.service;

import com.configcenter.model.App;
import com.configcenter.model.AuditAction;
import com.configcenter.model.AuditLog;
import com.configcenter.model.Config;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.Map;
import java.util.Optional;

@Service
@RequiredArgsConstructor
public class ConfigService {

    private final DataStore dataStore;
    private final AuditService auditService;

    public Optional<Config> getConfig(String appName, String key) {
        App app = dataStore.getApps().get(appName);
        if (app == null) {
            return Optional.empty();
        }
        return Optional.ofNullable(app.getConfigs().get(key));
    }

    public Map<String, Config> getAllConfigs(String appName) {
        App app = dataStore.getApps().get(appName);
        if (app == null) {
            return Map.of();
        }
        return app.getConfigs();
    }

    public Config createOrUpdateConfig(String appName, String key, String value, boolean secret, String userId) {
        App app = dataStore.getApps().get(appName);
        Config existing = app.getConfigs().get(key);
        
        String maskedValue = secret ? "***" : value;
        String oldValue = null;
        boolean isUpdate = existing != null;
        
        if (isUpdate) {
            oldValue = existing.isSecret() ? "***" : existing.getValue();
            existing.setValue(value);
            existing.setSecret(secret);
            existing.setUpdatedAt(Instant.now());
            existing.setUpdatedBy(userId);
            
            AuditLog log = auditService.createLog(userId, appName, AuditAction.UPDATE_CONFIG,
                    oldValue, maskedValue, key);
            auditService.saveLog(log);
            
            return existing;
        } else {
            Config config = new Config(key, value, secret, userId);
            app.getConfigs().put(key, config);
            
            AuditLog log = auditService.createLog(userId, appName, AuditAction.CREATE_CONFIG,
                    null, maskedValue, key);
            auditService.saveLog(log);
            
            return config;
        }
    }

    public boolean deleteConfig(String appName, String key, String userId) {
        App app = dataStore.getApps().get(appName);
        if (app == null) {
            return false;
        }
        
        Config removed = app.getConfigs().remove(key);
        if (removed != null) {
            String oldValue = removed.isSecret() ? "***" : removed.getValue();
            AuditLog log = auditService.createLog(userId, appName, AuditAction.DELETE_CONFIG,
                    oldValue, null, key);
            auditService.saveLog(log);
            return true;
        }
        return false;
    }

    public boolean configExists(String appName, String key) {
        App app = dataStore.getApps().get(appName);
        return app != null && app.getConfigs().containsKey(key);
    }
}
