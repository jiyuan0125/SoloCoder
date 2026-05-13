package com.example.gateway.model;

import java.util.Map;

public class RouteMatchResult {
    private final boolean matched;
    private final RouteDefinition route;
    private final Map<String, String> pathParams;
    private final String targetPath;

    public RouteMatchResult(boolean matched, RouteDefinition route, Map<String, String> pathParams, String targetPath) {
        this.matched = matched;
        this.route = route;
        this.pathParams = pathParams;
        this.targetPath = targetPath;
    }

    public boolean isMatched() {
        return matched;
    }

    public RouteDefinition getRoute() {
        return route;
    }

    public Map<String, String> getPathParams() {
        return pathParams;
    }

    public String getTargetPath() {
        return targetPath;
    }
}
