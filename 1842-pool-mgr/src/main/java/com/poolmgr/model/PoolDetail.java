package com.poolmgr.model;

public class PoolDetail {
    private PoolConfig config;
    private PoolStatus status;

    public PoolDetail() {}

    public PoolDetail(PoolConfig config, PoolStatus status) {
        this.config = config;
        this.status = status;
    }

    public PoolConfig getConfig() { return config; }
    public void setConfig(PoolConfig config) { this.config = config; }

    public PoolStatus getStatus() { return status; }
    public void setStatus(PoolStatus status) { this.status = status; }
}
