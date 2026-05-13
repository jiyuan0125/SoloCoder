package com.example.gateway;

import com.example.gateway.config.GatewayConfig;
import com.example.gateway.core.GatewayContext;
import com.example.gateway.core.GatewayServer;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class GatewayApplication {
    private static final Logger logger = LoggerFactory.getLogger(GatewayApplication.class);

    public static void main(String[] args) {
        GatewayConfig config = new GatewayConfig();
        GatewayContext context = new GatewayContext(config);
        GatewayServer server = new GatewayServer(context);

        logger.info("Starting version gateway on port {}...", config.getPort());
        logger.info("Default version: {}", config.getDefaultVersion());

        Runtime.getRuntime().addShutdownHook(new Thread(() -> {
            logger.info("Shutting down gateway...");
            server.shutdown();
        }));

        try {
            server.start();
        } catch (InterruptedException e) {
            logger.error("Gateway interrupted", e);
            Thread.currentThread().interrupt();
        }
    }
}
