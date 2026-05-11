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
public class SubmitAcceptanceRequest {

    @NotBlank(message = "操作员不能为空")
    private String operator;

    private String remark;
}
