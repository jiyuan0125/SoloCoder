package com.gateway.auth.model;

import java.time.Instant;

public class AuditLog {

    public enum Result {
        PASSED,
        REJECTED,
        EXPIRED,
        UPSTREAM_4XX
    }

    private final Instant timestamp;
    private final String clientId;
    private final String requestPath;
    private final Result result;

    public AuditLog(Instant timestamp, String clientId, String requestPath, Result result) {
        this.timestamp = timestamp;
        this.clientId = clientId;
        this.requestPath = requestPath;
        this.result = result;
    }

    public Instant getTimestamp() {
        return timestamp;
    }

    public String getClientId() {
        return clientId;
    }

    public String getRequestPath() {
        return requestPath;
    }

    public Result getResult() {
        return result;
    }
}
