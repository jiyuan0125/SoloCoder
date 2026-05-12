package com.example.serviceregistry.dto;

import jakarta.validation.constraints.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.Map;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class RegisterRequest {
    @NotBlank(message = "serviceName is required")
    private String serviceName;

    @NotBlank(message = "instanceId is required")
    private String instanceId;

    @NotBlank(message = "ip is required")
    @Pattern(
            regexp = "^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$",
            message = "Invalid IP address format"
    )
    private String ip;

    @NotNull(message = "port is required")
    @Min(value = 1, message = "Port must be >= 1")
    @Max(value = 65535, message = "Port must be <= 65535")
    private Integer port;

    @Builder.Default
    private Map<String, String> tags = Map.of();

    @Min(value = 0, message = "Weight must be >= 0")
    @Builder.Default
    private Integer weight = 1;

    @Min(value = 5, message = "Heartbeat interval must be >= 5 seconds")
    @Max(value = 300, message = "Heartbeat interval must be <= 300 seconds")
    @Builder.Default
    private Integer heartbeatInterval = 30;
}
