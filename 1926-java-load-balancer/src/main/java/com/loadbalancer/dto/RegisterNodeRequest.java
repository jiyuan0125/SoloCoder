package com.loadbalancer.dto;

import javax.validation.constraints.Min;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;

public class RegisterNodeRequest {

    @NotBlank(message = "ip is required")
    private String ip;

    @NotNull(message = "port is required")
    @Min(value = 1, message = "port must be positive")
    private Integer port;

    public String getIp() {
        return ip;
    }

    public void setIp(String ip) {
        this.ip = ip;
    }

    public Integer getPort() {
        return port;
    }

    public void setPort(Integer port) {
        this.port = port;
    }
}
