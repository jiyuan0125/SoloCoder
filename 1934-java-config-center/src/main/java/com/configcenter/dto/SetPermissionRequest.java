package com.configcenter.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class SetPermissionRequest {
    @NotBlank
    private String userId;
    @NotBlank
    private String role;
}
