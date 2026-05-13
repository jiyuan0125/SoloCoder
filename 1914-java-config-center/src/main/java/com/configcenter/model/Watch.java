package com.configcenter.model;

public class Watch {
    private String clientId;
    private String mode;
    private String callbackUrl;

    public Watch() {
    }

    public Watch(String clientId, String mode, String callbackUrl) {
        this.clientId = clientId;
        this.mode = mode;
        this.callbackUrl = callbackUrl;
    }

    public String getClientId() {
        return clientId;
    }

    public void setClientId(String clientId) {
        this.clientId = clientId;
    }

    public String getMode() {
        return mode;
    }

    public void setMode(String mode) {
        this.mode = mode;
    }

    public String getCallbackUrl() {
        return callbackUrl;
    }

    public void setCallbackUrl(String callbackUrl) {
        this.callbackUrl = callbackUrl;
    }
}
