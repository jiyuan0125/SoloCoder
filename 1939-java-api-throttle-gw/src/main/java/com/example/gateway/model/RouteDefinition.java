package com.example.gateway.model;

import java.util.List;
import java.util.Map;

public class RouteDefinition {
    private String path;
    private String targetUrl;
    private Map<String, String> targetHost;
    private String targetPath;
    private List<String> pathParamNames;

    public String getPath() {
        return path;
    }

    public void setPath(String path) {
        this.path = path;
    }

    public String getTargetUrl() {
        return targetUrl;
    }

    public void setTargetUrl(String targetUrl) {
        this.targetUrl = targetUrl;
    }

    public Map<String, String> getTargetHost() {
        return targetHost;
    }

    public void setTargetHost(Map<String, String> targetHost) {
        this.targetHost = targetHost;
    }

    public String getTargetPath() {
        return targetPath;
    }

    public void setTargetPath(String targetPath) {
        this.targetPath = targetPath;
    }

    public List<String> getPathParamNames() {
        return pathParamNames;
    }

    public void setPathParamNames(List<String> pathParamNames) {
        this.pathParamNames = pathParamNames;
    }
}
