package com.canary.router.model;

public class RoutingRule {
    public enum SourceType {
        HEADER, COOKIE, QUERY, IP
    }
    
    private SourceType source;
    private String key;
    private String value;
    private String targetVersion;
    
    public RoutingRule() {}
    
    public RoutingRule(SourceType source, String key, String value, String targetVersion) {
        this.source = source;
        this.key = key;
        this.value = value;
        this.targetVersion = targetVersion;
    }
    
    public SourceType getSource() { return source; }
    public void setSource(SourceType source) { this.source = source; }
    
    public String getKey() { return key; }
    public void setKey(String key) { this.key = key; }
    
    public String getValue() { return value; }
    public void setValue(String value) { this.value = value; }
    
    public String getTargetVersion() { return targetVersion; }
    public void setTargetVersion(String targetVersion) { this.targetVersion = targetVersion; }
}
