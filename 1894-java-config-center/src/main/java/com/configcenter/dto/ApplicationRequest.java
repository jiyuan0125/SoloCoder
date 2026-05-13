package com.configcenter.dto;

import lombok.Data;
import jakarta.validation.constraints.NotBlank;

@Data
public class ApplicationRequest {
    @NotBlank
    private String name;
    private String description;
}
