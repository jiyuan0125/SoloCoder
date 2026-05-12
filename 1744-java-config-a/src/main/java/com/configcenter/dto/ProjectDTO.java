package com.configcenter.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class ProjectDTO {
    @NotBlank(message = "Project name is required")
    private String name;

    private String description;

    private String validationCallbackUrl;
}
