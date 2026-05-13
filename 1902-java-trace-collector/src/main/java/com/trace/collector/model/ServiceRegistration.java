package com.trace.collector.model;

import com.fasterxml.jackson.annotation.JsonProperty;

public class ServiceRegistration {

    @JsonProperty("name")
    private String name;

    @JsonProperty("address")
    private String address;

    private long registeredAt;

    public ServiceRegistration() {
        this.registeredAt = System.currentTimeMillis();
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getAddress() {
        return address;
    }

    public void setAddress(String address) {
        this.address = address;
    }

    public long getRegisteredAt() {
        return registeredAt;
    }

    public void setRegisteredAt(long registeredAt) {
        this.registeredAt = registeredAt;
    }
}
