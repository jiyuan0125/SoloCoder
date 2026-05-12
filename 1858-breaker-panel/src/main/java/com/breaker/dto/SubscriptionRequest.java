package com.breaker.dto;

import com.fasterxml.jackson.annotation.JsonProperty;

import javax.validation.constraints.NotBlank;

public class SubscriptionRequest {
    @JsonProperty("service_name")
    @NotBlank(message = "service_name is required")
    private String serviceName;

    @JsonProperty("callback_url")
    @NotBlank(message = "callback_url is required")
    private String callbackUrl;

    public SubscriptionRequest() {}

    public String getServiceName() {
        return serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public String getCallbackUrl() {
        return callbackUrl;
    }

    public void setCallbackUrl(String callbackUrl) {
        this.callbackUrl = callbackUrl;
    }
}
