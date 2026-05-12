package com.canary.router;

import com.canary.router.config.ServerConfig;
import com.canary.router.netty.CanaryServer;

public class CanaryRouterApplication {
    public static void main(String[] args) {
        try {
            ServerConfig config = new ServerConfig();
            CanaryServer server = new CanaryServer(config);
            server.start();
        } catch (Exception e) {
            System.err.println("Failed to start Canary Router: " + e.getMessage());
            e.printStackTrace();
            System.exit(1);
        }
    }
}
