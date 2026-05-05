package com.training.common.dto.request;

public class CancelRegistrationRequest {
    private String registrationId;
    private String reason;

    public CancelRegistrationRequest() {
    }

    public String getRegistrationId() {
        return registrationId;
    }

    public void setRegistrationId(String registrationId) {
        this.registrationId = registrationId;
    }

    public String getReason() {
        return reason;
    }

    public void setReason(String reason) {
        this.reason = reason;
    }
}
