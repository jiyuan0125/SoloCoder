package com.healthchecker.dto;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class ServiceRegistrationRequest {

    @NotBlank(message = "服务名称不能为空")
    private String serviceName;

    @NotBlank(message = "检查URL不能为空")
    private String checkUrl;

    @Min(value = 1, message = "检查间隔至少为1秒")
    private int checkIntervalSeconds = 30;

    @Min(value = 1, message = "超时时间至少为1秒")
    private int timeoutSeconds = 5;

    private String callbackUrl;
}