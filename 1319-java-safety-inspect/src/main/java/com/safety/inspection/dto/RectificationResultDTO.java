package com.safety.inspection.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class RectificationResultDTO {

    @NotNull(message = "隐患ID不能为空")
    private Long hazardId;

    @NotBlank(message = "整改结果不能为空")
    private String rectificationResult;

    private String rectificationPhotos;
}
