package com.configcenter.service;

import com.configcenter.model.App;
import com.configcenter.model.Permission;
import org.springframework.stereotype.Component;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Component
public class DataStore {
    private final Map<String, App> apps = new ConcurrentHashMap<>();
    private final Map<String, Map<String, Permission>> permissions = new ConcurrentHashMap<>();

    public Map<String, App> getApps() {
        return apps;
    }

    public Map<String, Map<String, Permission>> getPermissions() {
        return permissions;
    }
}
