package com.gateway.engine;

import com.gateway.manager.VersionManager;
import com.gateway.model.Version;

import java.util.List;
import java.util.Map;
import java.util.Random;
import java.util.concurrent.ThreadLocalRandom;

public class TrafficDistributionEngine {
    private final VersionManager versionManager;

    public TrafficDistributionEngine(VersionManager versionManager) {
        this.versionManager = versionManager;
    }

    public String selectVersionByWeight() {
        Map<String, Version> versions = versionManager.getAllVersions();
        if (versions.isEmpty()) {
            return null;
        }

        int totalWeight = versions.values().stream()
                .mapToInt(Version::getWeight)
                .sum();

        if (totalWeight <= 0) {
            List<Version> versionList = versionManager.getVersionList();
            return versionList.get(0).getName();
        }

        Random random = ThreadLocalRandom.current();
        int randomValue = random.nextInt(totalWeight);

        int cumulative = 0;
        for (Version version : versions.values()) {
            cumulative += version.getWeight();
            if (randomValue < cumulative) {
                return version.getName();
            }
        }

        return versionManager.getVersionList().get(0).getName();
    }
}
