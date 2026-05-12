package com.poolmgr.model;

public class PoolStatus {
    private String name;
    private int activeConnections;
    private int idleConnections;
    private int waitingRequests;

    public PoolStatus() {}

    public PoolStatus(String name, int activeConnections, int idleConnections, int waitingRequests) {
        this.name = name;
        this.activeConnections = activeConnections;
        this.idleConnections = idleConnections;
        this.waitingRequests = waitingRequests;
    }

    public String getName() { return name; }
    public void setName(String name) { this.name = name; }

    public int getActiveConnections() { return activeConnections; }
    public void setActiveConnections(int activeConnections) { this.activeConnections = activeConnections; }

    public int getIdleConnections() { return idleConnections; }
    public void setIdleConnections(int idleConnections) { this.idleConnections = idleConnections; }

    public int getWaitingRequests() { return waitingRequests; }
    public void setWaitingRequests(int waitingRequests) { this.waitingRequests = waitingRequests; }
}
