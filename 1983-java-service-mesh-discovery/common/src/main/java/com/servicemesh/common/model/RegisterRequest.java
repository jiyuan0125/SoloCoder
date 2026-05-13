package com.servicemesh.common.model;

public class RegisterRequest {
    private String serviceName;
    private String ip;
    private int port;
    private String version;
    private String zone;

    public String getServiceName() { return serviceName; }
    public void setServiceName(String serviceName) { this.serviceName = serviceName; }
    public String getIp() { return ip; }
    public void setIp(String ip) { this.ip = ip; }
    public int getPort() { return port; }
    public void setPort(int port) { this.port = port; }
    public String getVersion() { return version; }
    public void setVersion(String version) { this.version = version; }
    public String getZone() { return zone; }
    public void setZone(String zone) { this.zone = zone; }
}
