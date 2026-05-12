package com.poolmgr.pool;

import com.poolmgr.model.DataSourceType;

import java.sql.Connection;
import java.sql.SQLException;
import java.time.Instant;
import java.util.UUID;
import redis.clients.jedis.Jedis;

public class ConnectionWrapper {
    private final String id;
    private final DataSourceType type;
    private final Connection sqlConnection;
    private final Jedis redisConnection;
    private Instant lastUsedTime;
    private volatile boolean closed;

    public ConnectionWrapper(DataSourceType type, Connection sqlConnection) {
        this.id = UUID.randomUUID().toString();
        this.type = type;
        this.sqlConnection = sqlConnection;
        this.redisConnection = null;
        this.lastUsedTime = Instant.now();
        this.closed = false;
    }

    public ConnectionWrapper(DataSourceType type, Jedis redisConnection) {
        this.id = UUID.randomUUID().toString();
        this.type = type;
        this.sqlConnection = null;
        this.redisConnection = redisConnection;
        this.lastUsedTime = Instant.now();
        this.closed = false;
    }

    public String getId() { return id; }
    public DataSourceType getType() { return type; }
    public Connection getSqlConnection() { return sqlConnection; }
    public Jedis getRedisConnection() { return redisConnection; }
    public Instant getLastUsedTime() { return lastUsedTime; }
    public void setLastUsedTime(Instant lastUsedTime) { this.lastUsedTime = lastUsedTime; }

    public boolean isValid() {
        if (closed) return false;
        try {
            if (type == DataSourceType.MYSQL && sqlConnection != null) {
                return sqlConnection.isValid(2);
            } else if (type == DataSourceType.REDIS && redisConnection != null) {
                return "PONG".equals(redisConnection.ping());
            }
        } catch (Exception e) {
            return false;
        }
        return false;
    }

    public void close() {
        if (closed) return;
        closed = true;
        try {
            if (sqlConnection != null) {
                sqlConnection.close();
            }
            if (redisConnection != null) {
                redisConnection.close();
            }
        } catch (SQLException e) {
        }
    }
}
