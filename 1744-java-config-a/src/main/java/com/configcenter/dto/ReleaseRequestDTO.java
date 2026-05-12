package com.configcenter.dto;

import com.configcenter.model.Release;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;

import java.util.List;

@Data
public class ReleaseRequestDTO {
    @NotBlank(message = "Environment is required")
    private String environment;

    private Release.ReleaseType type = Release.ReleaseType.FULL;

    private List<String> grayInstances;

    private String createdBy;
}
