package com.servicemesh.sidecar.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix = "sidecar")
public class SidecarProperties {
    private String registryUrl = "http://localhost:8761";
    private String controlPlaneUrl = "http://localhost:18000";
    private String serviceName = "unknown";
    private String appHost = "127.0.0.1";
    private int appPort = 8080;
    private String version = "v1";
    private String zone = "default";
    private String localIp = "127.0.0.1";
    private int proxyPort = 19001;
    private long heartbeatInterval = 5000;
    private long ruleRefreshInterval = 5000;

    public String getRegistryUrl() { return registryUrl; }
    public void setRegistryUrl(String registryUrl) { this.registryUrl = registryUrl; }
    public String getControlPlaneUrl() { return controlPlaneUrl; }
    public void setControlPlaneUrl(String controlPlaneUrl) { this.controlPlaneUrl = controlPlaneUrl; }
    public String getServiceName() { return serviceName; }
    public void setServiceName(String serviceName) { this.serviceName = serviceName; }
    public String getAppHost() { return appHost; }
    public void setAppHost(String appHost) { this.appHost = appHost; }
    public int getAppPort() { return appPort; }
    public void setAppPort(int appPort) { this.appPort = appPort; }
    public String getVersion() { return version; }
    public void setVersion(String version) { this.version = version; }
    public String getZone() { return zone; }
    public void setZone(String zone) { this.zone = zone; }
    public String getLocalIp() { return localIp; }
    public void setLocalIp(String localIp) { this.localIp = localIp; }
    public int getProxyPort() { return proxyPort; }
    public void setProxyPort(int proxyPort) { this.proxyPort = proxyPort; }
    public long getHeartbeatInterval() { return heartbeatInterval; }
    public void setHeartbeatInterval(long heartbeatInterval) { this.heartbeatInterval = heartbeatInterval; }
    public long getRuleRefreshInterval() { return ruleRefreshInterval; }
    public void setRuleRefreshInterval(long ruleRefreshInterval) { this.ruleRefreshInterval = ruleRefreshInterval; }
}
