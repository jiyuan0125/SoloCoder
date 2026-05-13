package com.example.registry.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class SubscribeRequest {
    @NotBlank(message = "Service name is required")
    private String serviceName;

    @NotBlank(message = "Callback URL is required")
    private String callbackUrl;
}
