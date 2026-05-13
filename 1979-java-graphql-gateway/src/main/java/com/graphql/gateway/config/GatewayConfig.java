package com.graphql.gateway.config;

import java.io.IOException;
import java.io.InputStream;
import java.util.Properties;

public class GatewayConfig {
    private static final String CONFIG_FILE = "application.properties";
    private static GatewayConfig instance;

    private int serverPort = 8080;
    private String serverHost = "0.0.0.0";
    private int maxQueryDepth = 5;
    private int maxFieldCount = 100;
    private int backendTimeoutSeconds = 30;

    private GatewayConfig() {
        loadConfig();
    }

    public static synchronized GatewayConfig getInstance() {
        if (instance == null) {
            instance = new GatewayConfig();
        }
        return instance;
    }

    private void loadConfig() {
        Properties props = new Properties();
        try (InputStream is = getClass().getClassLoader().getResourceAsStream(CONFIG_FILE)) {
            if (is != null) {
                props.load(is);
                serverPort = parseInt(props.getProperty("server.port"), serverPort);
                serverHost = props.getProperty("server.host", serverHost);
                maxQueryDepth = parseInt(props.getProperty("gateway.maxQueryDepth"), maxQueryDepth);
                maxFieldCount = parseInt(props.getProperty("gateway.maxFieldCount"), maxFieldCount);
                backendTimeoutSeconds = parseInt(props.getProperty("gateway.backend.timeoutSeconds"), backendTimeoutSeconds);
            }
        } catch (IOException e) {
            System.err.println("Failed to load config file: " + e.getMessage());
        }
    }

    private int parseInt(String value, int defaultValue) {
        if (value == null || value.trim().isEmpty()) {
            return defaultValue;
        }
        try {
            return Integer.parseInt(value.trim());
        } catch (NumberFormatException e) {
            return defaultValue;
        }
    }

    public int getServerPort() {
        return serverPort;
    }

    public String getServerHost() {
        return serverHost;
    }

    public int getMaxQueryDepth() {
        return maxQueryDepth;
    }

    public int getMaxFieldCount() {
        return maxFieldCount;
    }

    public int getBackendTimeoutSeconds() {
        return backendTimeoutSeconds;
    }
}
