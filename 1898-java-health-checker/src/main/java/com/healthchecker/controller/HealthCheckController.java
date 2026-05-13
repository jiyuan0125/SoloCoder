package com.healthchecker.controller;

import com.healthchecker.dto.ApiResponse;
import com.healthchecker.dto.ServiceRegistrationRequest;
import com.healthchecker.dto.ServiceStatusSummary;
import com.healthchecker.entity.CheckResult;
import com.healthchecker.entity.ServiceConfig;
import com.healthchecker.service.HealthCheckService;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/health")
public class HealthCheckController {

    private final HealthCheckService healthCheckService;

    public HealthCheckController(HealthCheckService healthCheckService) {
        this.healthCheckService = healthCheckService;
    }

    @PostMapping("/services")
    public ApiResponse<Void> registerService(@Valid @RequestBody ServiceRegistrationRequest request) {
        ServiceConfig config = new ServiceConfig();
        config.setServiceName(request.getServiceName());
        config.setCheckUrl(request.getCheckUrl());
        config.setCheckIntervalSeconds(request.getCheckIntervalSeconds());
        config.setTimeoutSeconds(request.getTimeoutSeconds());
        config.setCallbackUrl(request.getCallbackUrl());

        boolean registered = healthCheckService.registerService(config);

        if (registered) {
            return ApiResponse.success("服务注册成功", null);
        } else {
            return ApiResponse.error("服务已存在");
        }
    }

    @PostMapping("/services/{serviceName}/check")
    public ApiResponse<Void> triggerCheck(@PathVariable String serviceName) {
        boolean triggered = healthCheckService.triggerImmediateCheck(serviceName);
        if (triggered) {
            return ApiResponse.success("已触发立即检查", null);
        } else {
            return ApiResponse.error("服务不存在");
        }
    }

    @GetMapping("/services")
    public ApiResponse<List<ServiceStatusSummary>> getAllStatus() {
        List<ServiceStatusSummary> summaries = healthCheckService.getAllStatusSummary();
        return ApiResponse.success(summaries);
    }

    @GetMapping("/services/{serviceName}")
    public ApiResponse<ServiceConfig> getService(@PathVariable String serviceName) {
        ServiceConfig config = healthCheckService.getService(serviceName);
        if (config != null) {
            return ApiResponse.success(config);
        } else {
            return ApiResponse.error("服务不存在");
        }
    }

    @GetMapping("/services/{serviceName}/history")
    public ApiResponse<List<CheckResult>> getServiceHistory(@PathVariable String serviceName) {
        List<CheckResult> history = healthCheckService.getServiceHistory(serviceName);
        return ApiResponse.success(history);
    }

    @PostMapping("/services/{serviceName}/maintenance")
    public ApiResponse<Void> setMaintenance(@PathVariable String serviceName) {
        boolean success = healthCheckService.setMaintenance(serviceName);
        if (success) {
            return ApiResponse.success("服务已设置为维护中", null);
        } else {
            return ApiResponse.error("服务不存在");
        }
    }

    @DeleteMapping("/services/{serviceName}/maintenance")
    public ApiResponse<Void> cancelMaintenance(@PathVariable String serviceName) {
        boolean success = healthCheckService.cancelMaintenance(serviceName);
        if (success) {
            return ApiResponse.success("服务已取消维护", null);
        } else {
            return ApiResponse.error("服务不存在");
        }
    }
}