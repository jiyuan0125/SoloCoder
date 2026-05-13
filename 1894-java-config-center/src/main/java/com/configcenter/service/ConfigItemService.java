package com.configcenter.service;

import com.configcenter.constant.ConfigConstants;
import com.configcenter.dto.ConfigItemRequest;
import com.configcenter.dto.DiffResponse;
import com.configcenter.entity.ConfigItem;
import com.configcenter.entity.ConfigVersion;
import com.configcenter.exception.ResourceNotFoundException;
import com.configcenter.exception.ValidationException;
import com.configcenter.exception.ValueTooLargeException;
import com.configcenter.repository.ConfigItemRepository;
import com.configcenter.repository.ConfigVersionRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.nio.charset.StandardCharsets;
import java.util.*;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class ConfigItemService {

    private final ConfigItemRepository configItemRepository;
    private final ConfigVersionRepository configVersionRepository;
    private final ApplicationService applicationService;
    private final EnvironmentService environmentService;
    private final CallbackService callbackService;

    public ConfigItem getConfigItem(Long applicationId, Long environmentId, String configKey) {
        validateApplicationAndEnvironment(applicationId, environmentId);
        validateKeyName(configKey);
        return configItemRepository.findByApplicationIdAndEnvironmentIdAndConfigKey(applicationId, environmentId, configKey)
                .orElseThrow(() -> new ResourceNotFoundException(
                        "Config item not found for applicationId=" + applicationId + 
                        ", environmentId=" + environmentId + ", key=" + configKey));
    }

    public Page<ConfigItem> getConfigItemsByApplicationAndEnvironment(
            Long applicationId, Long environmentId, int offset, int limit) {
        validateApplicationAndEnvironment(applicationId, environmentId);
        if (limit <= 0 || limit > ConfigConstants.MAX_PAGE_SIZE) {
            limit = ConfigConstants.MAX_PAGE_SIZE;
        }
        int page = offset / limit;
        Pageable pageable = PageRequest.of(page, limit, Sort.by("configKey").ascending());
        return configItemRepository.findByApplicationIdAndEnvironmentId(applicationId, environmentId, pageable);
    }

    @Transactional
    public ConfigItem createConfigItem(Long applicationId, Long environmentId, ConfigItemRequest request) {
        validateApplicationAndEnvironment(applicationId, environmentId);
        validateKeyName(request.getConfigKey());
        validateValueSize(request.getConfigValue());
        
        if (configItemRepository.existsByApplicationIdAndEnvironmentIdAndConfigKey(
                applicationId, environmentId, request.getConfigKey())) {
            throw new ValidationException("Config key '" + request.getConfigKey() + "' already exists");
        }

        Long nextVersion = getNextVersionNumber(applicationId, environmentId);

        ConfigItem item = new ConfigItem();
        item.setApplicationId(applicationId);
        item.setEnvironmentId(environmentId);
        item.setConfigKey(request.getConfigKey());
        item.setConfigValue(request.getConfigValue());
        item.setVersionNumber(nextVersion);
        
        ConfigItem saved = configItemRepository.save(item);
        
        saveFullSnapshot(applicationId, environmentId, nextVersion, false);
        
        callbackService.notifyWatchersAsync(applicationId, environmentId, 
                request.getConfigKey(), null, request.getConfigValue());
        
        return saved;
    }

    @Transactional
    public ConfigItem updateConfigItem(Long applicationId, Long environmentId, String configKey, ConfigItemRequest request) {
        validateApplicationAndEnvironment(applicationId, environmentId);
        validateKeyName(configKey);
        validateKeyName(request.getConfigKey());
        validateValueSize(request.getConfigValue());

        ConfigItem existingItem = configItemRepository.findByApplicationIdAndEnvironmentIdAndConfigKey(
                applicationId, environmentId, configKey)
                .orElseThrow(() -> new ResourceNotFoundException(
                        "Config item not found for key=" + configKey));

        if (!configKey.equals(request.getConfigKey())) {
            if (configItemRepository.existsByApplicationIdAndEnvironmentIdAndConfigKey(
                    applicationId, environmentId, request.getConfigKey())) {
                throw new ValidationException("Config key '" + request.getConfigKey() + "' already exists");
            }
        }

        String oldValue = existingItem.getConfigValue();
        String newValue = request.getConfigValue();
        
        if (oldValue.equals(newValue) && configKey.equals(request.getConfigKey())) {
            return existingItem;
        }

        Long nextVersion = getNextVersionNumber(applicationId, environmentId);
        
        existingItem.setConfigKey(request.getConfigKey());
        existingItem.setConfigValue(newValue);
        existingItem.setVersionNumber(nextVersion);
        
        ConfigItem saved = configItemRepository.save(existingItem);
        
        saveFullSnapshot(applicationId, environmentId, nextVersion, false);
        
        callbackService.notifyWatchersAsync(applicationId, environmentId, 
                configKey, oldValue, newValue);
        
        return saved;
    }

    @Transactional
    public void deleteConfigItem(Long applicationId, Long environmentId, String configKey) {
        validateApplicationAndEnvironment(applicationId, environmentId);
        validateKeyName(configKey);
        
        ConfigItem existingItem = configItemRepository.findByApplicationIdAndEnvironmentIdAndConfigKey(
                applicationId, environmentId, configKey)
                .orElseThrow(() -> new ResourceNotFoundException(
                        "Config item not found for key=" + configKey));

        String oldValue = existingItem.getConfigValue();
        Long nextVersion = getNextVersionNumber(applicationId, environmentId);
        
        configItemRepository.delete(existingItem);
        
        saveFullSnapshot(applicationId, environmentId, nextVersion, false);
        
        callbackService.notifyWatchersAsync(applicationId, environmentId, 
                configKey, oldValue, null);
    }

    @Transactional
    public Long rollbackToVersion(Long applicationId, Long environmentId, Long targetVersion) {
        validateApplicationAndEnvironment(applicationId, environmentId);
        
        Long currentMaxVersion = configVersionRepository.findMaxVersionNumber(applicationId, environmentId)
                .orElse(0L);
        
        if (targetVersion < 1 || targetVersion > currentMaxVersion) {
            throw new ValidationException("Invalid target version: " + targetVersion + 
                    ". Must be between 1 and " + currentMaxVersion);
        }

        List<ConfigVersion> targetSnapshot = configVersionRepository
                .findByApplicationIdAndEnvironmentIdAndVersionNumber(applicationId, environmentId, targetVersion);
        
        if (targetSnapshot.isEmpty()) {
            throw new ResourceNotFoundException("No data found for version " + targetVersion);
        }

        Map<String, String> targetMap = targetSnapshot.stream()
                .filter(v -> v.getConfigValue() != null)
                .collect(Collectors.toMap(ConfigVersion::getConfigKey, ConfigVersion::getConfigValue));

        List<ConfigItem> currentItems = configItemRepository
                .findByApplicationIdAndEnvironmentId(applicationId, environmentId, 
                        Pageable.unpaged()).getContent();
        Map<String, ConfigItem> currentMap = currentItems.stream()
                .collect(Collectors.toMap(ConfigItem::getConfigKey, v -> v));

        Long nextVersion = currentMaxVersion + 1;

        Map<String, String> changes = new HashMap<>();

        for (Map.Entry<String, String> entry : targetMap.entrySet()) {
            String key = entry.getKey();
            String targetValue = entry.getValue();
            
            ConfigItem existing = currentMap.get(key);
            if (existing != null) {
                String oldValue = existing.getConfigValue();
                if (!oldValue.equals(targetValue)) {
                    existing.setConfigValue(targetValue);
                    existing.setVersionNumber(nextVersion);
                    configItemRepository.save(existing);
                    changes.put(key, oldValue);
                }
                currentMap.remove(key);
            } else {
                ConfigItem newItem = new ConfigItem();
                newItem.setApplicationId(applicationId);
                newItem.setEnvironmentId(environmentId);
                newItem.setConfigKey(key);
                newItem.setConfigValue(targetValue);
                newItem.setVersionNumber(nextVersion);
                configItemRepository.save(newItem);
                changes.put(key, null);
            }
        }

        for (ConfigItem remaining : currentMap.values()) {
            configItemRepository.delete(remaining);
            changes.put(remaining.getConfigKey(), remaining.getConfigValue());
        }

        saveFullSnapshot(applicationId, environmentId, nextVersion, true);

        for (Map.Entry<String, String> change : changes.entrySet()) {
            String key = change.getKey();
            String oldValue = change.getValue();
            String newValue = targetMap.get(key);
            callbackService.notifyWatchersAsync(applicationId, environmentId, key, oldValue, newValue);
        }

        return nextVersion;
    }

    public DiffResponse getDiffBetweenVersions(Long applicationId, Long environmentId, 
            Long fromVersion, Long toVersion) {
        validateApplicationAndEnvironment(applicationId, environmentId);
        
        if (fromVersion == null || toVersion == null) {
            throw new ValidationException("Both fromVersion and toVersion are required");
        }
        
        boolean swap = fromVersion > toVersion;
        Long actualFrom = swap ? toVersion : fromVersion;
        Long actualTo = swap ? fromVersion : toVersion;

        List<ConfigVersion> fromVersions = configVersionRepository
                .findByApplicationIdAndEnvironmentIdAndVersionNumber(applicationId, environmentId, actualFrom);
        List<ConfigVersion> toVersions = configVersionRepository
                .findByApplicationIdAndEnvironmentIdAndVersionNumber(applicationId, environmentId, actualTo);

        if (fromVersions.isEmpty() && toVersions.isEmpty()) {
            throw new ResourceNotFoundException("No data found for versions " + fromVersion + " or " + toVersion);
        }

        Map<String, String> fromMap = fromVersions.stream()
                .filter(v -> v.getConfigValue() != null)
                .collect(Collectors.toMap(ConfigVersion::getConfigKey, ConfigVersion::getConfigValue));
        Map<String, String> toMap = toVersions.stream()
                .filter(v -> v.getConfigValue() != null)
                .collect(Collectors.toMap(ConfigVersion::getConfigKey, ConfigVersion::getConfigValue));

        Set<String> allKeys = new HashSet<>();
        allKeys.addAll(fromMap.keySet());
        allKeys.addAll(toMap.keySet());

        List<DiffResponse.ConfigChange> changes = new ArrayList<>();
        
        for (String key : allKeys) {
            String oldValue = fromMap.get(key);
            String newValue = toMap.get(key);
            
            if (oldValue == null && newValue != null) {
                changes.add(new DiffResponse.ConfigChange(key, null, newValue, DiffResponse.ChangeType.ADDED));
            } else if (oldValue != null && newValue == null) {
                changes.add(new DiffResponse.ConfigChange(key, oldValue, null, DiffResponse.ChangeType.DELETED));
            } else if (oldValue != null && newValue != null && !oldValue.equals(newValue)) {
                changes.add(new DiffResponse.ConfigChange(key, oldValue, newValue, DiffResponse.ChangeType.MODIFIED));
            }
        }

        changes.sort(Comparator.comparing(DiffResponse.ConfigChange::getKey));

        return new DiffResponse(actualFrom, actualTo, changes);
    }

    private Long getNextVersionNumber(Long applicationId, Long environmentId) {
        return configVersionRepository.findMaxVersionNumber(applicationId, environmentId)
                .orElse(0L) + 1;
    }

    private void saveFullSnapshot(Long applicationId, Long environmentId, Long versionNumber, boolean isRollback) {
        List<ConfigItem> currentItems = configItemRepository
                .findByApplicationIdAndEnvironmentId(applicationId, environmentId, Pageable.unpaged()).getContent();
        
        List<ConfigVersion> versions = currentItems.stream()
                .map(item -> {
                    ConfigVersion v = new ConfigVersion();
                    v.setApplicationId(applicationId);
                    v.setEnvironmentId(environmentId);
                    v.setConfigKey(item.getConfigKey());
                    v.setConfigValue(item.getConfigValue());
                    v.setVersionNumber(versionNumber);
                    v.setIsRollback(isRollback);
                    return v;
                })
                .collect(Collectors.toList());
        
        configVersionRepository.saveAll(versions);
    }

    private void validateApplicationAndEnvironment(Long applicationId, Long environmentId) {
        if (!applicationService.existsById(applicationId)) {
            throw new ResourceNotFoundException("Application not found with id: " + applicationId);
        }
        if (!environmentService.existsById(environmentId)) {
            throw new ResourceNotFoundException("Environment not found with id: " + environmentId);
        }
    }

    private void validateKeyName(String key) {
        if (key == null || key.isEmpty()) {
            throw new ValidationException("Key name cannot be empty");
        }
        if (!key.matches(ConfigConstants.KEY_PATTERN)) {
            throw new ValidationException("Key name can only contain letters, numbers, underscores, and dots");
        }
    }

    private void validateValueSize(String value) {
        if (value == null) {
            return;
        }
        int size = value.getBytes(StandardCharsets.UTF_8).length;
        if (size > ConfigConstants.MAX_VALUE_SIZE_BYTES) {
            throw new ValueTooLargeException(
                    "Config value exceeds maximum size of " + ConfigConstants.MAX_VALUE_SIZE_KB + 
                    "KB (actual: " + (size / 1024) + "KB)");
        }
    }
}
