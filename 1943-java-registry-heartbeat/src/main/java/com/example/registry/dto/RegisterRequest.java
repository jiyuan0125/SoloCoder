package com.example.registry.dto;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class RegisterRequest {
    @NotBlank(message = "IP is required")
    private String ip;

    @NotNull(message = "Port is required")
    private Integer port;

    @Min(value = 0, message = "Weight must be >= 0")
    private Integer weight;

    @Min(value = 1, message = "Heartbeat TTL must be >= 1")
    @NotNull(message = "Heartbeat TTL is required")
    private Integer heartbeatTtlSeconds;
}
