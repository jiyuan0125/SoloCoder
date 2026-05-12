package com.example.ratelimiter.model;

public class BucketStats {
    private String path;
    private int currentQueueSize;
    private int capacity;
    private double rate;
    private long overflowCountLastMinute;

    public BucketStats(String path, int currentQueueSize, int capacity, double rate, long overflowCountLastMinute) {
        this.path = path;
        this.currentQueueSize = currentQueueSize;
        this.capacity = capacity;
        this.rate = rate;
        this.overflowCountLastMinute = overflowCountLastMinute;
    }

    public String getPath() {
        return path;
    }

    public int getCurrentQueueSize() {
        return currentQueueSize;
    }

    public int getCapacity() {
        return capacity;
    }

    public double getRate() {
        return rate;
    }

    public long getOverflowCountLastMinute() {
        return overflowCountLastMinute;
    }
}
