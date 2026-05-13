package com.example.gateway.ratelimit;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class RateLimitManager {
    private static final Logger logger = LoggerFactory.getLogger(RateLimitManager.class);
    private final Map<String, TokenBucket> buckets = new ConcurrentHashMap<>();

    public void configureRoute(String routePath, int capacity, int refillRate) {
        TokenBucket bucket = buckets.get(routePath);
        if (bucket == null) {
            bucket = new TokenBucket(capacity, refillRate);
            buckets.put(routePath, bucket);
            logger.info("Rate limit configured for route {}: capacity={}, refillRate={}", routePath, capacity, refillRate);
        } else {
            bucket.reconfigure(capacity, refillRate);
            logger.info("Rate limit reconfigured for route {}: capacity={}, refillRate={}", routePath, capacity, refillRate);
        }
    }

    public void removeRoute(String routePath) {
        buckets.remove(routePath);
        logger.info("Rate limit removed for route {}", routePath);
    }

    public boolean tryAcquire(String routePath) {
        TokenBucket bucket = buckets.get(routePath);
        if (bucket == null) {
            return true;
        }
        return bucket.tryAcquire();
    }

    public double getWaitTimeSeconds(String routePath) {
        TokenBucket bucket = buckets.get(routePath);
        if (bucket == null) {
            return 0;
        }
        return bucket.getWaitTimeSeconds();
    }

    public TokenBucket getBucket(String routePath) {
        return buckets.get(routePath);
    }

    public Map<String, TokenBucket> getAllBuckets() {
        return buckets;
    }
}
