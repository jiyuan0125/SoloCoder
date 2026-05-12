package com.gateway.model;

import java.util.concurrent.atomic.AtomicLong;

public class Stats {
    private String versionName;
    private AtomicLong requestCount;
    private AtomicLong errorCount;

    public Stats(String versionName) {
        this.versionName = versionName;
        this.requestCount = new AtomicLong(0);
        this.errorCount = new AtomicLong(0);
    }

    public String getVersionName() {
        return versionName;
    }

    public long getRequestCount() {
        return requestCount.get();
    }

    public long getErrorCount() {
        return errorCount.get();
    }

    public void incrementRequestCount() {
        requestCount.incrementAndGet();
    }

    public void incrementErrorCount() {
        errorCount.incrementAndGet();
    }
}
