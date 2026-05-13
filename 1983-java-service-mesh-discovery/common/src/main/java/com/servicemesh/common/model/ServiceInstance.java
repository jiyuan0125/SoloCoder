package com.servicemesh.common.model;

import com.fasterxml.jackson.annotation.JsonIgnore;

import java.util.Objects;

public class ServiceInstance {
    private String serviceName;
    private String instanceId;
    private String ip;
    private int port;
    private String version;
    private String zone;
    private long registerTime;
    private long lastHeartbeat;
    private boolean healthy;
    private int missedHeartbeats;

    public ServiceInstance() {}

    public ServiceInstance(String serviceName, String instanceId, String ip, int port, String version, String zone) {
        this.serviceName = serviceName;
        this.instanceId = instanceId;
        this.ip = ip;
        this.port = port;
        this.version = version;
        this.zone = zone;
        this.registerTime = System.currentTimeMillis();
        this.lastHeartbeat = System.currentTimeMillis();
        this.healthy = false;
        this.missedHeartbeats = 0;
    }

    @JsonIgnore
    public String getAddress() {
        return ip + ":" + port;
    }

    public String getServiceName() { return serviceName; }
    public void setServiceName(String serviceName) { this.serviceName = serviceName; }
    public String getInstanceId() { return instanceId; }
    public void setInstanceId(String instanceId) { this.instanceId = instanceId; }
    public String getIp() { return ip; }
    public void setIp(String ip) { this.ip = ip; }
    public int getPort() { return port; }
    public void setPort(int port) { this.port = port; }
    public String getVersion() { return version; }
    public void setVersion(String version) { this.version = version; }
    public String getZone() { return zone; }
    public void setZone(String zone) { this.zone = zone; }
    public long getRegisterTime() { return registerTime; }
    public void setRegisterTime(long registerTime) { this.registerTime = registerTime; }
    public long getLastHeartbeat() { return lastHeartbeat; }
    public void setLastHeartbeat(long lastHeartbeat) { this.lastHeartbeat = lastHeartbeat; }
    public boolean isHealthy() { return healthy; }
    public void setHealthy(boolean healthy) { this.healthy = healthy; }
    public int getMissedHeartbeats() { return missedHeartbeats; }
    public void setMissedHeartbeats(int missedHeartbeats) { this.missedHeartbeats = missedHeartbeats; }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        ServiceInstance that = (ServiceInstance) o;
        return Objects.equals(instanceId, that.instanceId);
    }

    @Override
    public int hashCode() {
        return Objects.hash(instanceId);
    }
}
