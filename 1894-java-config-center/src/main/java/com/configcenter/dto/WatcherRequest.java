package com.configcenter.dto;

import lombok.Data;
import jakarta.validation.constraints.NotBlank;

@Data
public class WatcherRequest {
    @NotBlank
    private String callbackUrl;
}
