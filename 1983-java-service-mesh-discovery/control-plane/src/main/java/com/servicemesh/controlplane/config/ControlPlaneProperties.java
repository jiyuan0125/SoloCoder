package com.servicemesh.controlplane.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix = "control-plane")
public class ControlPlaneProperties {
    private String registryUrl = "http://localhost:8761";
    private long topologyRefreshInterval = 10000;

    public String getRegistryUrl() { return registryUrl; }
    public void setRegistryUrl(String registryUrl) { this.registryUrl = registryUrl; }
    public long getTopologyRefreshInterval() { return topologyRefreshInterval; }
    public void setTopologyRefreshInterval(long topologyRefreshInterval) { this.topologyRefreshInterval = topologyRefreshInterval; }
}
