package com.gateway.model;

public class Rule {
    private String id;
    private RuleType type;
    private String key;
    private String value;
    private String targetVersion;

    public enum RuleType {
        HEADER,
        COOKIE,
        IP
    }

    public Rule() {
    }

    public Rule(String id, RuleType type, String key, String value, String targetVersion) {
        this.id = id;
        this.type = type;
        this.key = key;
        this.value = value;
        this.targetVersion = targetVersion;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public RuleType getType() {
        return type;
    }

    public void setType(RuleType type) {
        this.type = type;
    }

    public String getKey() {
        return key;
    }

    public void setKey(String key) {
        this.key = key;
    }

    public String getValue() {
        return value;
    }

    public void setValue(String value) {
        this.value = value;
    }

    public String getTargetVersion() {
        return targetVersion;
    }

    public void setTargetVersion(String targetVersion) {
        this.targetVersion = targetVersion;
    }
}
