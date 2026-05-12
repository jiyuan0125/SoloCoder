package com.gateway.manager;

import com.gateway.model.Version;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicReference;

public class VersionManager {
    private final AtomicReference<Map<String, Version>> versionsRef = new AtomicReference<>(new ConcurrentHashMap<>());

    public void addVersion(Version version) {
        Map<String, Version> current = new HashMap<>(versionsRef.get());
        current.put(version.getName(), version);
        versionsRef.set(new ConcurrentHashMap<>(current));
    }

    public boolean removeVersion(String name) {
        Map<String, Version> current = new HashMap<>(versionsRef.get());
        if (!current.containsKey(name)) {
            return false;
        }
        current.remove(name);
        versionsRef.set(new ConcurrentHashMap<>(current));
        return true;
    }

    public Version getVersion(String name) {
        return versionsRef.get().get(name);
    }

    public Map<String, Version> getAllVersions() {
        return versionsRef.get();
    }

    public boolean hasVersion(String name) {
        return versionsRef.get().containsKey(name);
    }

    public List<Version> getVersionList() {
        return new ArrayList<>(versionsRef.get().values());
    }
}
