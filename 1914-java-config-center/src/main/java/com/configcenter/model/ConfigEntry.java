package com.configcenter.model;

import java.time.Instant;

public class ConfigEntry {
    private String key;
    private String value;
    private boolean secret;
    private Instant createdAt;
    private Instant updatedAt;

    public ConfigEntry() {
    }

    public ConfigEntry(String key, String value, boolean secret, Instant createdAt, Instant updatedAt) {
        this.key = key;
        this.value = value;
        this.secret = secret;
        this.createdAt = createdAt;
        this.updatedAt = updatedAt;
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

    public boolean isSecret() {
        return secret;
    }

    public void setSecret(boolean secret) {
        this.secret = secret;
    }

    public Instant getCreatedAt() {
        return createdAt;
    }

    public void setCreatedAt(Instant createdAt) {
        this.createdAt = createdAt;
    }

    public Instant getUpdatedAt() {
        return updatedAt;
    }

    public void setUpdatedAt(Instant updatedAt) {
        this.updatedAt = updatedAt;
    }
}
