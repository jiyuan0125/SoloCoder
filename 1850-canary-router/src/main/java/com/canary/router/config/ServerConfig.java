package com.canary.router.config;

import java.io.IOException;
import java.io.InputStream;
import java.util.Properties;

public class ServerConfig {
    private static final String CONFIG_FILE = "application.properties";
    private static final String DEFAULT_PORT = "8080";
    
    private int port;
    private String stableHost;
    private int stablePort;
    private String canaryHost;
    private int canaryPort;
    
    public ServerConfig() {
        load();
    }
    
    private void load() {
        Properties props = new Properties();
        try (InputStream is = getClass().getClassLoader().getResourceAsStream(CONFIG_FILE)) {
            if (is != null) {
                props.load(is);
            }
        } catch (IOException e) {
            throw new RuntimeException("Failed to load configuration", e);
        }
        
        this.port = Integer.parseInt(System.getProperty("server.port", 
            props.getProperty("server.port", DEFAULT_PORT)));
        this.stableHost = props.getProperty("server.backend.stable.host", "127.0.0.1");
        this.stablePort = Integer.parseInt(props.getProperty("server.backend.stable.port", "8081"));
        this.canaryHost = props.getProperty("server.backend.canary.host", "127.0.0.1");
        this.canaryPort = Integer.parseInt(props.getProperty("server.backend.canary.port", "8082"));
    }
    
    public int getPort() { return port; }
    public String getStableHost() { return stableHost; }
    public int getStablePort() { return stablePort; }
    public String getCanaryHost() { return canaryHost; }
    public int getCanaryPort() { return canaryPort; }
}
