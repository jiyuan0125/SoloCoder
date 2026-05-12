package com.configcenter.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class ConfigItemDTO {
    @NotBlank(message = "Config key is required")
    private String configKey;

    private String value;
}
