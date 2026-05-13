package com.example.apiproxy.service;

import com.example.apiproxy.config.ProxyProperties;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpHeaders;
import org.springframework.stereotype.Service;

import java.io.Serializable;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class CacheService {

    private static final Logger logger = LoggerFactory.getLogger(CacheService.class);

    private final ProxyProperties properties;
    private final ConcurrentHashMap<String, CacheEntry> cache = new ConcurrentHashMap<>();

    public CacheService(ProxyProperties properties) {
        this.properties = properties;
    }

    public CacheEntry get(String key) {
        if (!properties.getCache().isEnabled()) {
            return null;
        }
        CacheEntry entry = cache.get(key);
        if (entry == null) {
            return null;
        }
        long now = System.currentTimeMillis();
        long ttlMillis = properties.getCache().getTtlSeconds() * 1000L;
        if (now - entry.timestamp > ttlMillis) {
            cache.remove(key);
            logger.debug("Cache expired for key: {}", key);
            return null;
        }
        logger.debug("Cache hit for key: {}", key);
        return entry;
    }

    public void put(String key, CacheEntry entry) {
        if (!properties.getCache().isEnabled()) {
            return;
        }
        cache.put(key, entry);
        logger.debug("Cache set for key: {}", key);
    }

    public String buildKey(String version, String path, String queryString) {
        StringBuilder sb = new StringBuilder();
        sb.append(version).append(":").append(path);
        if (queryString != null && !queryString.isEmpty()) {
            sb.append("?").append(queryString);
        }
        return sb.toString();
    }

    public static class CacheEntry implements Serializable {
        public final int statusCode;
        public final Map<String, List<String>> headers;
        public final byte[] body;
        public final long timestamp;

        public CacheEntry(int statusCode, HttpHeaders headers, byte[] body) {
            this.statusCode = statusCode;
            this.headers = headers;
            this.body = body;
            this.timestamp = System.currentTimeMillis();
        }
    }
}
