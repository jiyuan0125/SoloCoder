package com.example.ratelimiter.core;

import com.example.ratelimiter.model.QueuedRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.Queue;
import java.util.concurrent.ConcurrentLinkedDeque;
import java.util.concurrent.atomic.AtomicLong;

public class LeakyBucket {
    private static final Logger log = LoggerFactory.getLogger(LeakyBucket.class);

    private final String path;
    private volatile int capacity;
    private volatile double rate;
    private final Queue<QueuedRequest> queue;
    private final AtomicLong overflowCount;
    private final AtomicLong lastDrainTime;

    public LeakyBucket(String path, int capacity, double rate) {
        this.path = path;
        this.capacity = capacity;
        this.rate = rate;
        this.queue = new ConcurrentLinkedDeque<>();
        this.overflowCount = new AtomicLong(0);
        this.lastDrainTime = new AtomicLong(System.nanoTime());
    }

    public synchronized boolean tryEnqueue(QueuedRequest request) {
        if (queue.size() >= capacity) {
            overflowCount.incrementAndGet();
            return false;
        }
        queue.offer(request);
        log.debug("Enqueued request for path: {}, queue size: {}", path, queue.size());
        return true;
    }

    public QueuedRequest poll() {
        return queue.poll();
    }

    public int getCurrentQueueSize() {
        return queue.size();
    }

    public int getCapacity() {
        return capacity;
    }

    public void setCapacity(int capacity) {
        this.capacity = capacity;
    }

    public double getRate() {
        return rate;
    }

    public void setRate(double rate) {
        this.rate = rate;
    }

    public long getOverflowCountAndReset() {
        return overflowCount.getAndSet(0);
    }

    public long getLastDrainTime() {
        return lastDrainTime.get();
    }

    public void updateLastDrainTime(long time) {
        lastDrainTime.set(time);
    }

    public double getIntervalNanos() {
        return 1_000_000_000.0 / rate;
    }
}
