package com.example.gateway.ratelimit;

public class TokenBucket {
    private final int capacity;
    private final int refillRate;
    private long tokens;
    private long lastRefillTime;

    public TokenBucket(int capacity, int refillRate) {
        this.capacity = capacity;
        this.refillRate = refillRate;
        this.tokens = capacity;
        this.lastRefillTime = System.nanoTime();
    }

    public synchronized boolean tryAcquire() {
        refill();
        if (tokens > 0) {
            tokens--;
            return true;
        }
        return false;
    }

    public synchronized double getWaitTimeSeconds() {
        if (tokens > 0) {
            return 0;
        }
        return (1.0 / refillRate);
    }

    public synchronized void reconfigure(int capacity, int refillRate) {
        TokenBucket newBucket = new TokenBucket(capacity, refillRate);
        this.tokens = newBucket.tokens;
        this.lastRefillTime = newBucket.lastRefillTime;
    }

    private void refill() {
        long now = System.nanoTime();
        long elapsed = now - lastRefillTime;
        double secondsElapsed = elapsed / 1_000_000_000.0;
        long tokensToAdd = (long) (secondsElapsed * refillRate);

        if (tokensToAdd > 0) {
            tokens = Math.min(capacity, tokens + tokensToAdd);
            lastRefillTime = now;
        }
    }

    public int getCapacity() {
        return capacity;
    }

    public int getRefillRate() {
        return refillRate;
    }

    public synchronized long getTokens() {
        return tokens;
    }
}
