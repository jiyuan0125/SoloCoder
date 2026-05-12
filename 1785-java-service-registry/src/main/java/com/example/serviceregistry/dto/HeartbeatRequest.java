package com.example.serviceregistry.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class HeartbeatRequest {
    @NotBlank(message = "serviceName is required")
    private String serviceName;

    @NotBlank(message = "instanceId is required")
    private String instanceId;
}
