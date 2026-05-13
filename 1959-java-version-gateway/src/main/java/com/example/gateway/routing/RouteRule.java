package com.example.gateway.routing;

import java.util.List;

public class RouteRule {
    private String version;
    private List<String> backends;
    private boolean compatibilityEnabled;

    public RouteRule() {
    }

    public RouteRule(String version, List<String> backends, boolean compatibilityEnabled) {
        this.version = version;
        this.backends = backends;
        this.compatibilityEnabled = compatibilityEnabled;
    }

    public String getVersion() {
        return version;
    }

    public void setVersion(String version) {
        this.version = version;
    }

    public List<String> getBackends() {
        return backends;
    }

    public void setBackends(List<String> backends) {
        this.backends = backends;
    }

    public boolean isCompatibilityEnabled() {
        return compatibilityEnabled;
    }

    public void setCompatibilityEnabled(boolean compatibilityEnabled) {
        this.compatibilityEnabled = compatibilityEnabled;
    }

    public boolean isValid() {
        return version != null && !version.isEmpty()
                && backends != null && !backends.isEmpty();
    }
}
