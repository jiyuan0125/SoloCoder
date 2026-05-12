package com.logaggregate.model;

public class StoredLogEntry {

    private long timestampUtcMs;
    private String service;
    private String level;
    private String message;

    public StoredLogEntry(long timestampUtcMs, String service, String level, String message) {
        this.timestampUtcMs = timestampUtcMs;
        this.service = service;
        this.level = level.toUpperCase();
        this.message = message;
    }

    public long getTimestampUtcMs() {
        return timestampUtcMs;
    }

    public void setTimestampUtcMs(long timestampUtcMs) {
        this.timestampUtcMs = timestampUtcMs;
    }

    public String getService() {
        return service;
    }

    public void setService(String service) {
        this.service = service;
    }

    public String getLevel() {
        return level;
    }

    public void setLevel(String level) {
        this.level = level;
    }

    public String getMessage() {
        return message;
    }

    public void setMessage(String message) {
        this.message = message;
    }
}
