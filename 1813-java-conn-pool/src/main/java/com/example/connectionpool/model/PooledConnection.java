package com.example.connectionpool.model;

import lombok.Data;

import java.time.Instant;
import java.util.UUID;
import java.util.concurrent.atomic.AtomicLong;

@Data
public class PooledConnection {
    private static final AtomicLong ID_COUNTER = new AtomicLong(0);

    private final String id;
    private final String poolName;
    private volatile String borrowerThreadName;
    private volatile String borrowerStackTrace;
    private volatile Instant borrowedAt;
    private volatile boolean leaked;

    public PooledConnection(String poolName) {
        this.id = UUID.randomUUID().toString();
        this.poolName = poolName;
    }

    public void markBorrowed(String borrowerThreadName, String borrowerStackTrace) {
        this.borrowerThreadName = borrowerThreadName;
        this.borrowerStackTrace = borrowerStackTrace;
        this.borrowedAt = Instant.now();
        this.leaked = false;
    }

    public void markReturned() {
        this.borrowerThreadName = null;
        this.borrowerStackTrace = null;
        this.borrowedAt = null;
        this.leaked = false;
    }

    public long getBorrowedDurationSeconds() {
        if (borrowedAt == null) {
            return 0;
        }
        return Instant.now().getEpochSecond() - borrowedAt.getEpochSecond();
    }
}
