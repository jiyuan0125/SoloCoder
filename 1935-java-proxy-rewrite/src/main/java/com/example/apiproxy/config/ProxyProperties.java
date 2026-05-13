package com.example.apiproxy.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Component
@ConfigurationProperties(prefix = "proxy")
public class ProxyProperties {

    private int timeout = 10000;
    private String defaultVersion = "v1";
    private CacheConfig cache = new CacheConfig();
    private RetryConfig retry = new RetryConfig();
    private Map<String, List<String>> backends = new HashMap<>();

    public int getTimeout() {
        return timeout;
    }

    public void setTimeout(int timeout) {
        this.timeout = timeout;
    }

    public String getDefaultVersion() {
        return defaultVersion;
    }

    public void setDefaultVersion(String defaultVersion) {
        this.defaultVersion = defaultVersion;
    }

    public CacheConfig getCache() {
        return cache;
    }

    public void setCache(CacheConfig cache) {
        this.cache = cache;
    }

    public RetryConfig getRetry() {
        return retry;
    }

    public void setRetry(RetryConfig retry) {
        this.retry = retry;
    }

    public Map<String, List<String>> getBackends() {
        return backends;
    }

    public void setBackends(Map<String, List<String>> backends) {
        this.backends = backends;
    }

    public List<String> getBackends(String version) {
        return backends.getOrDefault(version, new ArrayList<>());
    }

    public static class CacheConfig {
        private boolean enabled = true;
        private int ttlSeconds = 5;

        public boolean isEnabled() {
            return enabled;
        }

        public void setEnabled(boolean enabled) {
            this.enabled = enabled;
        }

        public int getTtlSeconds() {
            return ttlSeconds;
        }

        public void setTtlSeconds(int ttlSeconds) {
            this.ttlSeconds = ttlSeconds;
        }
    }

    public static class RetryConfig {
        private int maxAttempts = 2;

        public int getMaxAttempts() {
            return maxAttempts;
        }

        public void setMaxAttempts(int maxAttempts) {
            this.maxAttempts = maxAttempts;
        }
    }
}
