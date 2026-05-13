package com.example.gateway.config;

import com.example.gateway.model.GatewayConfig;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.io.InputStream;
import java.util.Properties;

public class ConfigLoader {
    private static final Logger logger = LoggerFactory.getLogger(ConfigLoader.class);
    private static final String CONFIG_FILE = "application.properties";

    public static GatewayConfig load() {
        GatewayConfig config = new GatewayConfig();
        Properties props = new Properties();

        try (InputStream is = ConfigLoader.class.getClassLoader().getResourceAsStream(CONFIG_FILE)) {
            if (is != null) {
                props.load(is);
                String port = props.getProperty("server.port");
                if (port != null) {
                    config.setServerPort(Integer.parseInt(port.trim()));
                }
                String timeout = props.getProperty("http.client.timeout.seconds");
                if (timeout != null) {
                    config.setClientTimeoutSeconds(Integer.parseInt(timeout.trim()));
                }
                logger.info("Configuration loaded from {}: port={}, timeout={}s",
                        CONFIG_FILE, config.getServerPort(), config.getClientTimeoutSeconds());
            } else {
                logger.warn("{} not found, using defaults", CONFIG_FILE);
            }
        } catch (IOException e) {
            logger.error("Failed to load configuration, using defaults", e);
        }

        return config;
    }
}
