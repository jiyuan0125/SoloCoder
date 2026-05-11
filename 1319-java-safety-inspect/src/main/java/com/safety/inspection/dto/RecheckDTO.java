package com.safety.inspection.dto;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class RecheckDTO {

    @NotNull(message = "隐患ID不能为空")
    private Long hazardId;

    @NotBlank(message = "复检结果不能为空")
    private String recheckResult;

    private String recheckRemark;
}
