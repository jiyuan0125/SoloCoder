package com.configcenter.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class CreateOrUpdateConfigRequest {
    private String key;
    @NotBlank
    private String value;
    private boolean secret = false;
}
