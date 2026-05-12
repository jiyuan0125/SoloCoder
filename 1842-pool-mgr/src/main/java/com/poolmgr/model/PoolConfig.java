package com.poolmgr.model;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;

public class PoolConfig {

    @NotBlank
    private String name;

    @NotNull
    private DataSourceType type;

    @NotBlank
    private String host;

    @NotNull
    @Min(1)
    private Integer port;

    private String username;

    private String password;

    private String database;

    @NotNull
    @Min(1)
    private Integer maxConnections;

    @NotNull
    @Min(0)
    private Integer minIdle;

    @NotNull
    @Min(1)
    private Integer timeoutSeconds;

    public String getName() { return name; }
    public void setName(String name) { this.name = name; }

    public DataSourceType getType() { return type; }
    public void setType(DataSourceType type) { this.type = type; }

    public String getHost() { return host; }
    public void setHost(String host) { this.host = host; }

    public Integer getPort() { return port; }
    public void setPort(Integer port) { this.port = port; }

    public String getUsername() { return username; }
    public void setUsername(String username) { this.username = username; }

    public String getPassword() { return password; }
    public void setPassword(String password) { this.password = password; }

    public String getDatabase() { return database; }
    public void setDatabase(String database) { this.database = database; }

    public Integer getMaxConnections() { return maxConnections; }
    public void setMaxConnections(Integer maxConnections) { this.maxConnections = maxConnections; }

    public Integer getMinIdle() { return minIdle; }
    public void setMinIdle(Integer minIdle) { this.minIdle = minIdle; }

    public Integer getTimeoutSeconds() { return timeoutSeconds; }
    public void setTimeoutSeconds(Integer timeoutSeconds) { this.timeoutSeconds = timeoutSeconds; }
}
