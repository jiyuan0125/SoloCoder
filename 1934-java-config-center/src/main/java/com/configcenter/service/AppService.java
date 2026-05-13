package com.configcenter.service;

import com.configcenter.model.App;
import com.configcenter.model.AuditAction;
import com.configcenter.model.AuditLog;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.Map;
import java.util.Optional;

@Service
@RequiredArgsConstructor
public class AppService {

    private final DataStore dataStore;
    private final AuditService auditService;

    public App createApp(String name, String description, String userId) {
        App app = new App(name, description);
        dataStore.getApps().put(name, app);
        
        AuditLog log = auditService.createLog(userId, name, AuditAction.CREATE_APP,
                null, "App created: " + name, null);
        auditService.saveLog(log);
        
        return app;
    }

    public Optional<App> getApp(String name) {
        return Optional.ofNullable(dataStore.getApps().get(name));
    }

    public boolean appExists(String name) {
        return dataStore.getApps().containsKey(name);
    }

    public void deleteApp(String name, String userId) {
        App app = dataStore.getApps().remove(name);
        if (app != null) {
            dataStore.getPermissions().remove(name);
            
            AuditLog log = auditService.createLog(userId, name, AuditAction.DELETE_APP,
                    "App deleted: " + name, null, null);
            auditService.saveLog(log);
        }
    }

    public Map<String, App> getAllApps() {
        return dataStore.getApps();
    }
}
