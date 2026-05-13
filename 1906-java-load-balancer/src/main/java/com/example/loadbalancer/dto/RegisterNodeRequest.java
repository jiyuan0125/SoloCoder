package com.example.loadbalancer.dto;

import lombok.Data;

import javax.validation.constraints.Min;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Positive;

@Data
public class RegisterNodeRequest {
    @NotBlank(message = "IP is required")
    private String ip;

    @NotNull(message = "Port is required")
    @Positive(message = "Port must be positive")
    private Integer port;

    @Min(value = 1, message = "Weight must be at least 1")
    private int weight = 1;
}
