package com.example.gateway.config;

import java.io.IOException;
import java.io.InputStream;
import java.util.Properties;

public class GatewayConfig {
    private static final String CONFIG_FILE = "application.properties";
    private final Properties properties;

    public GatewayConfig() {
        this.properties = loadProperties();
    }

    private Properties loadProperties() {
        Properties props = new Properties();
        try (InputStream is = Thread.currentThread()
                .getContextClassLoader()
                .getResourceAsStream(CONFIG_FILE)) {
            if (is != null) {
                props.load(is);
            }
        } catch (IOException e) {
            e.printStackTrace();
        }
        return props;
    }

    public int getPort() {
        String value = properties.getProperty("gateway.port", "8080");
        return Integer.parseInt(value);
    }

    public String getDefaultVersion() {
        return properties.getProperty("gateway.default.version", "v2");
    }
}
