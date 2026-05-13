package com.example.cachemiddleware.cache;

import com.example.cachemiddleware.config.CacheProperties;
import com.example.cachemiddleware.model.*;
import com.example.cachemiddleware.notify.SubscriptionManager;
import com.example.cachemiddleware.stats.StatsCollector;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.nio.charset.StandardCharsets;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.locks.ReentrantReadWriteLock;

@Slf4j
@Service
public class CacheManager {

    private final Map<String, NamespaceContext> namespaces = new ConcurrentHashMap<>();
    private final CacheProperties cacheProperties;
    private final StatsCollector statsCollector;
    private final SubscriptionManager subscriptionManager;
    private final ObjectMapper objectMapper = new ObjectMapper();

    public CacheManager(CacheProperties cacheProperties,
                        StatsCollector statsCollector,
                        SubscriptionManager subscriptionManager) {
        this.cacheProperties = cacheProperties;
        this.statsCollector = statsCollector;
        this.subscriptionManager = subscriptionManager;
    }

    private NamespaceContext getOrCreateContext(String namespace) {
        return namespaces.computeIfAbsent(namespace, this::createDefaultContext);
    }

    private NamespaceContext createDefaultContext(String namespace) {
        CacheProperties.DefaultConfig defaultConfig = cacheProperties.getDefaultConfig();
        return new NamespaceContext(
                NamespaceConfig.builder()
                        .namespace(namespace)
                        .ttlSeconds(defaultConfig.getTtlSeconds())
                        .maxCapacity(defaultConfig.getMaxCapacity())
                        .evictionPolicy(defaultConfig.getEvictionPolicy())
                        .build()
        );
    }

    public void updateConfig(String namespace, NamespaceConfig newConfig) {
        NamespaceContext ctx = getOrCreateContext(namespace);
        ctx.lock.writeLock().lock();
        try {
            NamespaceConfig current = ctx.config;
            NamespaceConfig updated = NamespaceConfig.builder()
                    .namespace(namespace)
                    .ttlSeconds(newConfig.getTtlSeconds() > 0 ? newConfig.getTtlSeconds() : current.getTtlSeconds())
                    .maxCapacity(newConfig.getMaxCapacity() > 0 ? newConfig.getMaxCapacity() : current.getMaxCapacity())
                    .evictionPolicy(newConfig.getEvictionPolicy() != null ? newConfig.getEvictionPolicy() : current.getEvictionPolicy())
                    .build();
            if (updated.getEvictionPolicy() != current.getEvictionPolicy()) {
                CacheStrategy newStrategy = createStrategy(updated.getEvictionPolicy());
                for (CacheEntry entry : ctx.strategy.getAll()) {
                    newStrategy.put(entry);
                }
                ctx.strategy = newStrategy;
            }
            if (updated.getMaxCapacity() < ctx.strategy.size()) {
                while (ctx.strategy.size() > updated.getMaxCapacity()) {
                    CacheEntry evicted = ctx.strategy.evict();
                    if (evicted != null) {
                        statsCollector.recordEviction(namespace);
                        publishChange(namespace, evicted.getKey(), "EVICT", evicted.getValue(), null);
                    }
                }
            }
            ctx.config = updated;
            log.info("Config updated: namespace={}, ttl={}, capacity={}, policy={}",
                    namespace, updated.getTtlSeconds(), updated.getMaxCapacity(), updated.getEvictionPolicy());
        } finally {
            ctx.lock.writeLock().unlock();
        }
    }

    public NamespaceConfig getConfig(String namespace) {
        NamespaceContext ctx = getOrCreateContext(namespace);
        return ctx.config;
    }

    public Object get(String namespace, String key) {
        NamespaceContext ctx = getOrCreateContext(namespace);
        ctx.lock.readLock().lock();
        try {
            CacheEntry entry = ctx.strategy.get(key);
            if (entry == null) {
                statsCollector.recordMiss(namespace);
                return null;
            }
            if (entry.isExpired()) {
                ctx.lock.readLock().unlock();
                ctx.lock.writeLock().lock();
                try {
                    ctx.strategy.remove(key);
                } finally {
                    ctx.lock.writeLock().unlock();
                    ctx.lock.readLock().lock();
                }
                statsCollector.recordMiss(namespace);
                publishChange(namespace, key, "EXPIRED", entry.getValue(), null);
                return null;
            }
            ctx.strategy.onAccess(entry);
            statsCollector.recordHit(namespace);
            return entry.getValue();
        } finally {
            ctx.lock.readLock().unlock();
        }
    }

    public boolean put(String namespace, String key, Object value) {
        return put(namespace, key, value, null);
    }

    public boolean put(String namespace, String key, Object value, Long ttlSecondsOverride) {
        if (valueSizeExceedsLimit(value)) {
            return false;
        }
        NamespaceContext ctx = getOrCreateContext(namespace);
        ctx.lock.writeLock().lock();
        try {
            if (ctx.strategy.get(key) != null) {
                return false;
            }
            ensureCapacity(ctx, namespace);
            long ttl = ttlSecondsOverride != null ? ttlSecondsOverride : ctx.config.getTtlSeconds();
            CacheEntry entry = new CacheEntry(key, value, ttl);
            ctx.strategy.put(entry);
            publishChange(namespace, key, "PUT", null, value);
            return true;
        } finally {
            ctx.lock.writeLock().unlock();
        }
    }

    public boolean update(String namespace, String key, Object value) {
        return update(namespace, key, value, null);
    }

    public boolean update(String namespace, String key, Object value, Long ttlSecondsOverride) {
        if (valueSizeExceedsLimit(value)) {
            return false;
        }
        NamespaceContext ctx = getOrCreateContext(namespace);
        ctx.lock.writeLock().lock();
        try {
            CacheEntry existing = ctx.strategy.get(key);
            if (existing == null) {
                return false;
            }
            long ttl = ttlSecondsOverride != null ? ttlSecondsOverride : ctx.config.getTtlSeconds();
            Object oldValue = existing.getValue();
            CacheEntry newEntry = new CacheEntry(key, value, ttl);
            ctx.strategy.put(newEntry);
            publishChange(namespace, key, "UPDATE", oldValue, value);
            return true;
        } finally {
            ctx.lock.writeLock().unlock();
        }
    }

    public boolean delete(String namespace, String key) {
        NamespaceContext ctx = getOrCreateContext(namespace);
        ctx.lock.writeLock().lock();
        try {
            CacheEntry removed = ctx.strategy.remove(key);
            if (removed != null) {
                publishChange(namespace, key, "DELETE", removed.getValue(), null);
                return true;
            }
            return false;
        } finally {
            ctx.lock.writeLock().unlock();
        }
    }

    public void clearNamespace(String namespace) {
        NamespaceContext ctx = namespaces.get(namespace);
        if (ctx == null) {
            return;
        }
        ctx.lock.writeLock().lock();
        try {
            for (CacheEntry entry : new ArrayList<>(ctx.strategy.getAll())) {
                ctx.strategy.remove(entry.getKey());
                publishChange(namespace, entry.getKey(), "DELETE", entry.getValue(), null);
            }
            statsCollector.reset(namespace);
        } finally {
            ctx.lock.writeLock().unlock();
        }
    }

    public BatchWriteResult batchPut(String namespace, Map<String, Object> entries) {
        NamespaceContext ctx = getOrCreateContext(namespace);
        ctx.lock.writeLock().lock();
        try {
            for (Map.Entry<String, Object> entry : entries.entrySet()) {
                if (valueSizeExceedsLimit(entry.getValue())) {
                    return BatchWriteResult.fail("Value size exceeds limit for key: " + entry.getKey());
                }
            }
            for (Map.Entry<String, Object> entry : entries.entrySet()) {
                if (ctx.strategy.get(entry.getKey()) != null) {
                    return BatchWriteResult.fail("Key already exists: " + entry.getKey());
                }
            }
            for (Map.Entry<String, Object> entry : entries.entrySet()) {
                ensureCapacity(ctx, namespace);
                CacheEntry cacheEntry = new CacheEntry(entry.getKey(), entry.getValue(), ctx.config.getTtlSeconds());
                ctx.strategy.put(cacheEntry);
                publishChange(namespace, entry.getKey(), "PUT", null, entry.getValue());
            }
            return BatchWriteResult.success();
        } finally {
            ctx.lock.writeLock().unlock();
        }
    }

    public NamespaceStats getStats(String namespace) {
        NamespaceContext ctx = getOrCreateContext(namespace);
        return NamespaceStats.builder()
                .namespace(namespace)
                .hitCount(statsCollector.getHitCount(namespace))
                .missCount(statsCollector.getMissCount(namespace))
                .evictionCount(statsCollector.getEvictionCount(namespace))
                .currentCapacity(ctx.strategy.size())
                .ttlSeconds(ctx.config.getTtlSeconds())
                .maxCapacity(ctx.config.getMaxCapacity())
                .evictionPolicy(ctx.config.getEvictionPolicy())
                .build();
    }

    private void ensureCapacity(NamespaceContext ctx, String namespace) {
        while (ctx.strategy.size() >= ctx.config.getMaxCapacity()) {
            CacheEntry evicted = ctx.strategy.evict();
            if (evicted != null) {
                statsCollector.recordEviction(namespace);
                publishChange(namespace, evicted.getKey(), "EVICT", evicted.getValue(), null);
            }
        }
    }

    private boolean valueSizeExceedsLimit(Object value) {
        if (value == null) {
            return false;
        }
        try {
            String json = objectMapper.writeValueAsString(value);
            int size = json.getBytes(StandardCharsets.UTF_8).length;
            return size > cacheProperties.getMaxValueSizeBytes();
        } catch (Exception e) {
            return true;
        }
    }

    private void publishChange(String namespace, String key, String changeType, Object oldValue, Object newValue) {
        KeyChangeEvent event = new KeyChangeEvent(namespace, key, changeType, oldValue, newValue, System.currentTimeMillis());
        subscriptionManager.notifyChange(event);
    }

    private CacheStrategy createStrategy(EvictionPolicy policy) {
        switch (policy) {
            case LFU:
                return new LfuCacheStrategy();
            case LRU:
            default:
                return new LruCacheStrategy();
        }
    }

    public static class NamespaceContext {
        volatile NamespaceConfig config;
        volatile CacheStrategy strategy;
        final ReentrantReadWriteLock lock = new ReentrantReadWriteLock();

        public NamespaceContext(NamespaceConfig config) {
            this.config = config;
            this.strategy = new LruCacheStrategy();
        }
    }

    public static class BatchWriteResult {
        public final boolean success;
        public final String errorMessage;

        private BatchWriteResult(boolean success, String errorMessage) {
            this.success = success;
            this.errorMessage = errorMessage;
        }

        public static BatchWriteResult success() {
            return new BatchWriteResult(true, null);
        }

        public static BatchWriteResult fail(String message) {
            return new BatchWriteResult(false, message);
        }
    }
}
