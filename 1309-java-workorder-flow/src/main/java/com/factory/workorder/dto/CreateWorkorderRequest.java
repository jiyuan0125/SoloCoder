package com.factory.workorder.dto;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CreateWorkorderRequest {

    @NotBlank(message = "设备名称不能为空")
    private String equipmentName;

    @NotBlank(message = "故障描述不能为空")
    private String faultDescription;

    @NotBlank(message = "报修人不能为空")
    private String reporter;

    @NotNull(message = "优先级不能为空")
    private String priority;
}
