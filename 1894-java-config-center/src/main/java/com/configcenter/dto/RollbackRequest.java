package com.configcenter.dto;

import lombok.Data;
import jakarta.validation.constraints.NotNull;

@Data
public class RollbackRequest {
    @NotNull
    private Long targetVersion;
}
