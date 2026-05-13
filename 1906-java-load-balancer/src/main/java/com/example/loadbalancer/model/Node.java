package com.example.loadbalancer.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;
import lombok.Builder;

import java.util.concurrent.atomic.AtomicInteger;

@Data
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class Node {
    private String id;
    private String ip;
    private int port;
    private int initialWeight;
    private int currentWeight;
    private NodeStatus status;
    private AtomicInteger activeConnections;
    private long registeredAt;
    private long lastHealthCheckAt;
    private boolean lastHealthCheckPassed;

    public Node(String id, String ip, int port, int initialWeight) {
        this.id = id;
        this.ip = ip;
        this.port = port;
        this.initialWeight = initialWeight;
        this.currentWeight = initialWeight;
        this.status = NodeStatus.PENDING_VERIFICATION;
        this.activeConnections = new AtomicInteger(0);
        this.registeredAt = System.currentTimeMillis();
    }

    public void markDegraded() {
        if (this.status == NodeStatus.HEALTHY || this.status == NodeStatus.PENDING_VERIFICATION) {
            this.status = NodeStatus.DEGRADED;
            this.currentWeight = Math.max(1, this.initialWeight / 2);
        }
    }

    public void markHealthy() {
        this.status = NodeStatus.HEALTHY;
        this.currentWeight = this.initialWeight;
    }

    public void markOffline() {
        this.status = NodeStatus.OFFLINE;
    }

    public void markPending() {
        this.status = NodeStatus.PENDING_VERIFICATION;
        this.currentWeight = this.initialWeight;
    }

    public void incrementConnections() {
        this.activeConnections.incrementAndGet();
    }

    public void decrementConnections() {
        this.activeConnections.updateAndGet(current -> Math.max(0, current - 1));
    }

    public int getActiveConnectionsCount() {
        return this.activeConnections.get();
    }

    public boolean canReceiveTraffic() {
        return this.status == NodeStatus.HEALTHY || this.status == NodeStatus.DEGRADED;
    }

    public String getAddress() {
        return this.ip + ":" + this.port;
    }
}
