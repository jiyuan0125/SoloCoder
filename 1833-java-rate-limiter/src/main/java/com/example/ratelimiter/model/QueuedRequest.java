package com.example.ratelimiter.model;

import jakarta.servlet.http.HttpServletRequest;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.atomic.AtomicBoolean;

public class QueuedRequest {
    private final HttpServletRequest request;
    private final CompletableFuture<Void> completionFuture;
    private final AtomicBoolean processed = new AtomicBoolean(false);
    private final int capacity;
    private final double rate;

    public QueuedRequest(HttpServletRequest request, int capacity, double rate) {
        this.request = request;
        this.completionFuture = new CompletableFuture<>();
        this.capacity = capacity;
        this.rate = rate;
    }

    public HttpServletRequest getRequest() {
        return request;
    }

    public CompletableFuture<Void> getCompletionFuture() {
        return completionFuture;
    }

    public int getCapacity() {
        return capacity;
    }

    public double getRate() {
        return rate;
    }

    public void markProcessed() {
        if (processed.compareAndSet(false, true)) {
            completionFuture.complete(null);
        }
    }

    public boolean isProcessed() {
        return processed.get();
    }
}
