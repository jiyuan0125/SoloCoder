package com.example.serviceregistry.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class SubscribeRequest {
    @NotBlank(message = "subscriberId is required")
    private String subscriberId;

    @NotBlank(message = "serviceName is required")
    private String serviceName;

    @NotBlank(message = "callbackUrl is required")
    @Pattern(
            regexp = "^https?://.*",
            message = "Invalid callback URL format, must start with http:// or https://"
    )
    private String callbackUrl;
}
