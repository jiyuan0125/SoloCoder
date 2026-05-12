package com.configcenter.controller;

import com.configcenter.dto.ApiResponse;
import com.configcenter.dto.ConfigItemDTO;
import com.configcenter.model.ConfigItem;
import com.configcenter.service.ConfigService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/projects/{projectId}/configs")
@RequiredArgsConstructor
public class ConfigController {

    private final ConfigService configService;

    @PostMapping("/{environment}")
    public ApiResponse<ConfigItem> setConfig(
            @PathVariable Long projectId,
            @PathVariable String environment,
            @Valid @RequestBody ConfigItemDTO dto) {
        ConfigItem config = configService.setConfig(projectId, environment, dto);
        return ApiResponse.success("Config updated successfully", config);
    }

    @GetMapping("/{environment}/{configKey}")
    public ApiResponse<ConfigItem> getConfig(
            @PathVariable Long projectId,
            @PathVariable String environment,
            @PathVariable String configKey) {
        ConfigItem config = configService.getConfig(projectId, environment, configKey);
        return ApiResponse.success(config);
    }

    @GetMapping("/{environment}")
    public ApiResponse<List<ConfigItem>> getConfigsByEnvironment(
            @PathVariable Long projectId,
            @PathVariable String environment) {
        List<ConfigItem> configs = configService.getConfigsByEnvironment(projectId, environment);
        return ApiResponse.success(configs);
    }

    @GetMapping("/{environment}/pending")
    public ApiResponse<List<ConfigItem>> getPendingConfigs(
            @PathVariable Long projectId,
            @PathVariable String environment) {
        List<ConfigItem> configs = configService.getPendingConfigs(projectId, environment);
        return ApiResponse.success(configs);
    }

    @GetMapping("/{environment}/released")
    public ApiResponse<List<ConfigItem>> getReleasedConfigs(
            @PathVariable Long projectId,
            @PathVariable String environment) {
        List<ConfigItem> configs = configService.getReleasedConfigs(projectId, environment);
        return ApiResponse.success(configs);
    }

    @DeleteMapping("/{environment}/{configKey}")
    public ApiResponse<Void> deleteConfig(
            @PathVariable Long projectId,
            @PathVariable String environment,
            @PathVariable String configKey) {
        configService.deleteConfig(projectId, environment, configKey);
        return ApiResponse.success("Config deleted successfully", null);
    }
}
