package com.canary.router.model;

import java.util.concurrent.atomic.AtomicLong;

public class Stats {
    private String versionName;
    private final AtomicLong requestCount = new AtomicLong(0);
    private final AtomicLong errorCount = new AtomicLong(0);
    
    public Stats(String versionName) {
        this.versionName = versionName;
    }
    
    public String getVersionName() { return versionName; }
    public void setVersionName(String versionName) { this.versionName = versionName; }
    
    public long getRequestCount() { return requestCount.get(); }
    public void incrementRequest() { requestCount.incrementAndGet(); }
    
    public long getErrorCount() { return errorCount.get(); }
    public void incrementError() { errorCount.incrementAndGet(); }
}
