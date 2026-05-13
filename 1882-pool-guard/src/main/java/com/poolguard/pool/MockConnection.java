package com.poolguard.pool;

import java.util.UUID;

public class MockConnection {
    private final String id;
    private boolean closed;
    private boolean valid;

    public MockConnection() {
        this.id = UUID.randomUUID().toString();
        this.closed = false;
        this.valid = true;
    }

    public String getId() {
        return id;
    }

    public boolean isClosed() {
        return closed;
    }

    public void close() {
        this.closed = true;
    }

    public boolean isValid() {
        return valid && !closed;
    }

    public void setValid(boolean valid) {
        this.valid = valid;
    }
}
