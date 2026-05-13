package com.example.gateway.model;

public class GatewayConfig {
    private int serverPort;
    private int clientTimeoutSeconds;

    public GatewayConfig() {
        this.serverPort = 8080;
        this.clientTimeoutSeconds = 5;
    }

    public int getServerPort() {
        return serverPort;
    }

    public void setServerPort(int serverPort) {
        this.serverPort = serverPort;
    }

    public int getClientTimeoutSeconds() {
        return clientTimeoutSeconds;
    }

    public void setClientTimeoutSeconds(int clientTimeoutSeconds) {
        this.clientTimeoutSeconds = clientTimeoutSeconds;
    }
}
