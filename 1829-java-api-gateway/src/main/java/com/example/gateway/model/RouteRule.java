package com.example.gateway.model;

import java.util.List;

public class RouteRule {
    private String id;
    private String pathPrefix;
    private List<BackendTarget> backends;
    private List<String> filterNames;
    private int weight;

    public RouteRule() {}

    public RouteRule(String id, String pathPrefix, List<BackendTarget> backends, List<String> filterNames, int weight) {
        this.id = id;
        this.pathPrefix = pathPrefix;
        this.backends = backends;
        this.filterNames = filterNames;
        this.weight = weight;
    }

    public String getId() { return id; }
    public void setId(String id) { this.id = id; }

    public String getPathPrefix() { return pathPrefix; }
    public void setPathPrefix(String pathPrefix) { this.pathPrefix = pathPrefix; }

    public List<BackendTarget> getBackends() { return backends; }
    public void setBackends(List<BackendTarget> backends) { this.backends = backends; }

    public List<String> getFilterNames() { return filterNames; }
    public void setFilterNames(List<String> filterNames) { this.filterNames = filterNames; }

    public int getWeight() { return weight; }
    public void setWeight(int weight) { this.weight = weight; }
}
