package com.configcenter.dto;

import lombok.Data;
import jakarta.validation.constraints.NotBlank;

@Data
public class ConfigItemRequest {
    @NotBlank
    private String configKey;
    @NotBlank
    private String configValue;
}
