package com.logaggregate.model;

public class RetentionConfig {
    private String service;
    private int retentionDays;

    public RetentionConfig(String service, int retentionDays) {
        this.service = service;
        this.retentionDays = retentionDays;
    }

    public String getService() {
        return service;
    }

    public void setService(String service) {
        this.service = service;
    }

    public int getRetentionDays() {
        return retentionDays;
    }

    public void setRetentionDays(int retentionDays) {
        this.retentionDays = retentionDays;
    }
}
