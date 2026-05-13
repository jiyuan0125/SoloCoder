package com.configcenter.controller;

import com.configcenter.dto.ConfigItemRequest;
import com.configcenter.dto.DiffResponse;
import com.configcenter.dto.RollbackRequest;
import com.configcenter.entity.ConfigItem;
import com.configcenter.service.ConfigItemService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/api/applications/{applicationId}/environments/{environmentId}/configs")
@RequiredArgsConstructor
public class ConfigItemController {

    private final ConfigItemService configItemService;

    @GetMapping
    public ResponseEntity<Page<ConfigItem>> getConfigItems(
            @PathVariable Long applicationId,
            @PathVariable Long environmentId,
            @RequestParam(defaultValue = "0") int offset,
            @RequestParam(defaultValue = "100") int limit) {
        return ResponseEntity.ok(configItemService.getConfigItemsByApplicationAndEnvironment(
                applicationId, environmentId, offset, limit));
    }

    @GetMapping("/{configKey}")
    public ResponseEntity<ConfigItem> getConfigItem(
            @PathVariable Long applicationId,
            @PathVariable Long environmentId,
            @PathVariable String configKey) {
        return ResponseEntity.ok(configItemService.getConfigItem(applicationId, environmentId, configKey));
    }

    @PostMapping
    public ResponseEntity<ConfigItem> createConfigItem(
            @PathVariable Long applicationId,
            @PathVariable Long environmentId,
            @Valid @RequestBody ConfigItemRequest request) {
        ConfigItem created = configItemService.createConfigItem(applicationId, environmentId, request);
        return new ResponseEntity<>(created, HttpStatus.CREATED);
    }

    @PutMapping("/{configKey}")
    public ResponseEntity<ConfigItem> updateConfigItem(
            @PathVariable Long applicationId,
            @PathVariable Long environmentId,
            @PathVariable String configKey,
            @Valid @RequestBody ConfigItemRequest request) {
        return ResponseEntity.ok(configItemService.updateConfigItem(applicationId, environmentId, configKey, request));
    }

    @DeleteMapping("/{configKey}")
    public ResponseEntity<Void> deleteConfigItem(
            @PathVariable Long applicationId,
            @PathVariable Long environmentId,
            @PathVariable String configKey) {
        configItemService.deleteConfigItem(applicationId, environmentId, configKey);
        return ResponseEntity.noContent().build();
    }

    @PostMapping("/rollback")
    public ResponseEntity<Map<String, Object>> rollback(
            @PathVariable Long applicationId,
            @PathVariable Long environmentId,
            @Valid @RequestBody RollbackRequest request) {
        Long newVersion = configItemService.rollbackToVersion(applicationId, environmentId, request.getTargetVersion());
        Map<String, Object> response = new HashMap<>();
        response.put("message", "Rollback successful");
        response.put("newVersion", newVersion);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/diff")
    public ResponseEntity<DiffResponse> getDiff(
            @PathVariable Long applicationId,
            @PathVariable Long environmentId,
            @RequestParam Long fromVersion,
            @RequestParam Long toVersion) {
        return ResponseEntity.ok(configItemService.getDiffBetweenVersions(
                applicationId, environmentId, fromVersion, toVersion));
    }
}
