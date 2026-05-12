package com.example.connectionpool.exception;

public class PoolFullException extends PoolException {
    private final int queueSize;

    public PoolFullException(String poolName, int queueSize) {
        super(poolName, String.format("连接池 %s 已满，当前队列长度 %d", poolName, queueSize));
        this.queueSize = queueSize;
    }

    public int getQueueSize() {
        return queueSize;
    }
}
