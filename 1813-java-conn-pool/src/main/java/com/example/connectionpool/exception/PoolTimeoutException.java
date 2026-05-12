package com.example.connectionpool.exception;

public class PoolTimeoutException extends PoolException {
    public PoolTimeoutException(String poolName) {
        super(poolName, String.format("获取连接超时，池名 %s", poolName));
    }
}
