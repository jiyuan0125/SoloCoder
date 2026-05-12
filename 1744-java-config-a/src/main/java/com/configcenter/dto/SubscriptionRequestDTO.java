package com.configcenter.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class SubscriptionRequestDTO {
    @NotBlank(message = "Instance ID is required")
    private String instanceId;

    @NotBlank(message = "Environment is required")
    private String environment;

    private String lastKnownVersion;
}
