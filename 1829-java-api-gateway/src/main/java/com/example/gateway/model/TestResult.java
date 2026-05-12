package com.example.gateway.model;

import java.util.List;

public class TestResult {
    private String matchedRouteId;
    private String matchedPathPrefix;
    private String backendUrl;
    private List<String> filterChain;

    public TestResult() {}

    public TestResult(String matchedRouteId, String matchedPathPrefix, String backendUrl, List<String> filterChain) {
        this.matchedRouteId = matchedRouteId;
        this.matchedPathPrefix = matchedPathPrefix;
        this.backendUrl = backendUrl;
        this.filterChain = filterChain;
    }

    public String getMatchedRouteId() { return matchedRouteId; }
    public void setMatchedRouteId(String matchedRouteId) { this.matchedRouteId = matchedRouteId; }

    public String getMatchedPathPrefix() { return matchedPathPrefix; }
    public void setMatchedPathPrefix(String matchedPathPrefix) { this.matchedPathPrefix = matchedPathPrefix; }

    public String getBackendUrl() { return backendUrl; }
    public void setBackendUrl(String backendUrl) { this.backendUrl = backendUrl; }

    public List<String> getFilterChain() { return filterChain; }
    public void setFilterChain(List<String> filterChain) { this.filterChain = filterChain; }
}
