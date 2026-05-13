package com.loadbalancer.dto;

import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class RegisterNodeRequest {
    @NotNull(message = "Address is required")
    private String address;
    private Integer weight = 1;
}
