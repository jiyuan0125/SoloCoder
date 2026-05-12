package com.example.connectionpool.exception;

import lombok.Getter;

@Getter
public class PoolException extends RuntimeException {
    private final String poolName;

    public PoolException(String poolName, String message) {
        super(message);
        this.poolName = poolName;
    }
}
