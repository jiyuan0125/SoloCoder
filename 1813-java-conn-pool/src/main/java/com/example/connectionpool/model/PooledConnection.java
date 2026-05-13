package com.example.connectionpool.model;

import lombok.Data;
import lombok.EqualsAndHashCode;
import lombok.ToString;

import java.time.Instant;
import java.util.UUID;
import java.util.concurrent.atomic.AtomicLong;

@Data
@EqualsAndHashCode(onlyExplicitlyIncluded = true)
@ToString(onlyExplicitlyIncluded = true)
public class PooledConnection {

    @EqualsAndHashCode.Include
    @ToString.Include
    private final String id;

    @EqualsAndHashCode.Include
    @ToString.Include
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
