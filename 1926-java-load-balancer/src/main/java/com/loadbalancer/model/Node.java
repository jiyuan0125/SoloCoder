package com.loadbalancer.model;

import java.time.Instant;

public class Node {

    private String ip;
    private int port;
    private NodeStatus status;
    private int weight;
    private Instant registeredAt;
    private Instant statusChangedAt;

    public Node() {
    }

    public Node(String ip, int port) {
        this.ip = ip;
        this.port = port;
        this.status = NodeStatus.REGISTERING;
        this.weight = 1;
        this.registeredAt = Instant.now();
        this.statusChangedAt = Instant.now();
    }

    public String getIp() {
        return ip;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public int getPort() {
        return port;
    }

    public void setPort(int port) {
        this.port = port;
    }

    public NodeStatus getStatus() {
        return status;
    }

    public void setStatus(NodeStatus status) {
        this.status = status;
        this.statusChangedAt = Instant.now();
    }

    public int getWeight() {
        return weight;
    }

    public void setWeight(int weight) {
        this.weight = weight;
    }

    public Instant getRegisteredAt() {
        return registeredAt;
    }

    public void setRegisteredAt(Instant registeredAt) {
        this.registeredAt = registeredAt;
    }

    public Instant getStatusChangedAt() {
        return statusChangedAt;
    }

    public void setStatusChangedAt(Instant statusChangedAt) {
        this.statusChangedAt = statusChangedAt;
    }

    public String getKey() {
        return ip + ":" + port;
    }

    public boolean canReceiveTraffic() {
        return status == NodeStatus.ACTIVE;
    }

    public boolean canReceiveExistingSessions() {
        return status == NodeStatus.ACTIVE || status == NodeStatus.DRAINING;
    }
}
