package com.example.gateway.model;

import java.util.Map;

public class FilterConfig {
    private String routeId;
    private String filterName;
    private Map<String, Object> properties;

    public FilterConfig() {}

    public FilterConfig(String routeId, String filterName, Map<String, Object> properties) {
        this.routeId = routeId;
        this.filterName = filterName;
        this.properties = properties;
    }

    public String getRouteId() { return routeId; }
    public void setRouteId(String routeId) { this.routeId = routeId; }

    public String getFilterName() { return filterName; }
    public void setFilterName(String filterName) { this.filterName = filterName; }

    public Map<String, Object> getProperties() { return properties; }
    public void setProperties(Map<String, Object> properties) { this.properties = properties; }
}
