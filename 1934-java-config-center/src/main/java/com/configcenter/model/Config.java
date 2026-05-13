package com.configcenter.model;

import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class Config {
    @NotBlank
    private String key;
    @NotBlank
    private String value;
    private boolean secret = false;
    private Instant createdAt;
    private Instant updatedAt;
    private String createdBy;
    private String updatedBy;

    public Config(String key, String value, boolean secret, String createdBy) {
        this.key = key;
        this.value = value;
        this.secret = secret;
        this.createdAt = Instant.now();
        this.updatedAt = Instant.now();
        this.createdBy = createdBy;
        this.updatedBy = createdBy;
    }
}
