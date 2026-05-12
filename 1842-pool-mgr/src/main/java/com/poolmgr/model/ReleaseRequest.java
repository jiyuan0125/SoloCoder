package com.poolmgr.model;

import jakarta.validation.constraints.NotBlank;

public class ReleaseRequest {

    @NotBlank
    private String connectionId;

    public String getConnectionId() { return connectionId; }
    public void setConnectionId(String connectionId) { this.connectionId = connectionId; }
}
