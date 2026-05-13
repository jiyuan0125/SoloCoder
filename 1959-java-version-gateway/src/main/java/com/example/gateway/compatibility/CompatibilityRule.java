package com.example.gateway.compatibility;

import java.util.List;

public class CompatibilityRule {
    private String sourceVersion;
    private String targetVersion;
    private List<FieldMapping> requestMappings;
    private List<FieldMapping> responseMappings;

    public CompatibilityRule() {
    }

    public String getSourceVersion() {
        return sourceVersion;
    }

    public void setSourceVersion(String sourceVersion) {
        this.sourceVersion = sourceVersion;
    }

    public String getTargetVersion() {
        return targetVersion;
    }

    public void setTargetVersion(String targetVersion) {
        this.targetVersion = targetVersion;
    }

    public List<FieldMapping> getRequestMappings() {
        return requestMappings;
    }

    public void setRequestMappings(List<FieldMapping> requestMappings) {
        this.requestMappings = requestMappings;
    }

    public List<FieldMapping> getResponseMappings() {
        return responseMappings;
    }

    public void setResponseMappings(List<FieldMapping> responseMappings) {
        this.responseMappings = responseMappings;
    }
}
