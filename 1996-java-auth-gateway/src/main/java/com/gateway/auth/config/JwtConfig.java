package com.gateway.auth.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import javax.annotation.PostConstruct;
import java.time.Instant;
import java.time.format.DateTimeParseException;

@Component
@ConfigurationProperties(prefix = "jwt")
public class JwtConfig {

    private String primarySecret;
    private String oldSecret;
    private int oldKeyExpireMinutes = 60;
    private String oldKeyExpireAt;

    private Instant computedOldKeyExpireAt;

    @PostConstruct
    public void init() {
        if (oldKeyExpireAt != null && !oldKeyExpireAt.trim().isEmpty()) {
            try {
                computedOldKeyExpireAt = Instant.parse(oldKeyExpireAt);
            } catch (DateTimeParseException e) {
                computedOldKeyExpireAt = null;
            }
        }
    }

    public String getPrimarySecret() {
        return primarySecret;
    }

    public void setPrimarySecret(String primarySecret) {
        this.primarySecret = primarySecret;
    }

    public String getOldSecret() {
        return oldSecret;
    }

    public void setOldSecret(String oldSecret) {
        this.oldSecret = oldSecret;
    }

    public int getOldKeyExpireMinutes() {
        return oldKeyExpireMinutes;
    }

    public void setOldKeyExpireMinutes(int oldKeyExpireMinutes) {
        this.oldKeyExpireMinutes = oldKeyExpireMinutes;
    }

    public String getOldKeyExpireAt() {
        return oldKeyExpireAt;
    }

    public void setOldKeyExpireAt(String oldKeyExpireAt) {
        this.oldKeyExpireAt = oldKeyExpireAt;
    }

    public Instant getComputedOldKeyExpireAt() {
        return computedOldKeyExpireAt;
    }
}
