package com.example.pool.core;

import lombok.Getter;

import java.time.Instant;

public class PooledConnection {
    
    @Getter
    private final Object connection;
    
    @Getter
    private final Instant createdAt;
    
    @Getter
    private Instant lastReturnedAt;
    
    public PooledConnection(Object connection) {
        this.connection = connection;
        this.createdAt = Instant.now();
        this.lastReturnedAt = Instant.now();
    }
    
    public void markReturned() {
        this.lastReturnedAt = Instant.now();
    }
}
