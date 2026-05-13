package com.configcenter.dto;

import lombok.Data;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;

@Data
public class EnvironmentRequest {
    @NotBlank
    @Pattern(regexp = "^(dev|staging|prod)$", message = "Environment must be dev, staging, or prod")
    private String name;
    private String description;
}
