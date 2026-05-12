package com.poolmgr.model;

public class AcquiredConnection {
    private String connectionId;
    private String poolName;
    private DataSourceType type;

    public AcquiredConnection() {}

    public AcquiredConnection(String connectionId, String poolName, DataSourceType type) {
        this.connectionId = connectionId;
        this.poolName = poolName;
        this.type = type;
    }

    public String getConnectionId() { return connectionId; }
    public void setConnectionId(String connectionId) { this.connectionId = connectionId; }

    public String getPoolName() { return poolName; }
    public void setPoolName(String poolName) { this.poolName = poolName; }

    public DataSourceType getType() { return type; }
    public void setType(DataSourceType type) { this.type = type; }
}
