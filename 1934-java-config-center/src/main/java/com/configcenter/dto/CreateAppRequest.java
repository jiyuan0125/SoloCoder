package com.configcenter.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class CreateAppRequest {
    @NotBlank
    private String name;
    private String description;
}
