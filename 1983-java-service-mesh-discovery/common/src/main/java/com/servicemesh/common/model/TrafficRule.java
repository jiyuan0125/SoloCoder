package com.servicemesh.common.model;

import java.util.List;
import java.util.Map;

public class TrafficRule {
    private String serviceName;
    private String type;
    private long lastUpdate;
    private VersionMatch versionMatch;
    private WeightBased weightBased;
    private ZoneBased zoneBased;

    public String getServiceName() { return serviceName; }
    public void setServiceName(String serviceName) { this.serviceName = serviceName; }
    public String getType() { return type; }
    public void setType(String type) { this.type = type; }
    public long getLastUpdate() { return lastUpdate; }
    public void setLastUpdate(long lastUpdate) { this.lastUpdate = lastUpdate; }
    public VersionMatch getVersionMatch() { return versionMatch; }
    public void setVersionMatch(VersionMatch versionMatch) { this.versionMatch = versionMatch; }
    public WeightBased getWeightBased() { return weightBased; }
    public void setWeightBased(WeightBased weightBased) { this.weightBased = weightBased; }
    public ZoneBased getZoneBased() { return zoneBased; }
    public void setZoneBased(ZoneBased zoneBased) { this.zoneBased = zoneBased; }

    public static class VersionMatch {
        private String targetVersion;

        public String getTargetVersion() { return targetVersion; }
        public void setTargetVersion(String targetVersion) { this.targetVersion = targetVersion; }
    }

    public static class WeightBased {
        private List<VersionWeight> weights;

        public List<VersionWeight> getWeights() { return weights; }
        public void setWeights(List<VersionWeight> weights) { this.weights = weights; }
    }

    public static class VersionWeight {
        private String version;
        private int weight;

        public String getVersion() { return version; }
        public void setVersion(String version) { this.version = version; }
        public int getWeight() { return weight; }
        public void setWeight(int weight) { this.weight = weight; }
    }

    public static class ZoneBased {
        private String preferZone;
        private boolean fallbackEnabled;

        public String getPreferZone() { return preferZone; }
        public void setPreferZone(String preferZone) { this.preferZone = preferZone; }
        public boolean isFallbackEnabled() { return fallbackEnabled; }
        public void setFallbackEnabled(boolean fallbackEnabled) { this.fallbackEnabled = fallbackEnabled; }
    }
}
