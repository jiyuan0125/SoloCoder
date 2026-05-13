package com.configcenter.dto;

import lombok.Data;

@Data
public class ConfigResponse {
    private String key;
    private String value;
    private boolean secret;

    public ConfigResponse(String key, String value, boolean secret) {
        this.key = key;
        this.value = secret ? "***" : value;
        this.secret = secret;
    }
}
