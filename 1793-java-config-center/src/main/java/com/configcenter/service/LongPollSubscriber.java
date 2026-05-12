package com.configcenter.service;

import lombok.Getter;

import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

public class LongPollSubscriber {
    
    @Getter
    private final String namespace;
    
    @Getter
    private final String group;
    
    @Getter
    private final String key;
    
    private final CountDownLatch latch = new CountDownLatch(1);
    
    private volatile boolean changed = false;
    
    public LongPollSubscriber(String namespace, String group, String key) {
        this.namespace = namespace;
        this.group = group;
        this.key = key;
    }
    
    public void onChange(String namespace, String group, String key) {
        this.changed = true;
        latch.countDown();
    }
    
    public boolean waitForChange(long timeout, TimeUnit unit) throws InterruptedException {
        boolean completed = latch.await(timeout, unit);
        return completed && changed;
    }
    
    public boolean isChanged() {
        return changed;
    }
}
