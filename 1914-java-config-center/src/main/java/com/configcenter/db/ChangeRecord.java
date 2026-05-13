package com.configcenter.db;

import java.time.Instant;

public class ChangeRecord {
    private final int id;
    private final String key;
    private final String action;
    private final String oldValue;
    private final String newValue;
    private final boolean isSecret;
    private final Instant timestamp;

    public ChangeRecord(int id, String key, String action, String oldValue, String newValue, boolean isSecret, Instant timestamp) {
        this.id = id;
        this.key = key;
        this.action = action;
        this.oldValue = oldValue;
        this.newValue = newValue;
        this.isSecret = isSecret;
        this.timestamp = timestamp;
    }

    public int getId() {
        return id;
    }

    public String getKey() {
        return key;
    }

    public String getAction() {
        return action;
    }

    public String getOldValue() {
        return oldValue;
    }

    public String getNewValue() {
        return newValue;
    }

    public boolean isSecret() {
        return isSecret;
    }

    public Instant getTimestamp() {
        return timestamp;
    }
}
