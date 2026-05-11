package com.factory.workorder.dto;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import javax.validation.constraints.NotBlank;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AssignWorkorderRequest {

    @NotBlank(message = "操作员不能为空")
    private String operator;

    @NotBlank(message = "当前处理人不能为空")
    private String currentHandler;

    private String remark;
}
