package com.configcenter.service;

import com.configcenter.dto.ConfigItemDTO;
import com.configcenter.model.ConfigItem;
import com.configcenter.model.Project;
import com.configcenter.repository.ConfigItemRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Slf4j
@Service
@RequiredArgsConstructor
public class ConfigService {

    private final ConfigItemRepository configItemRepository;
    private final ProjectService projectService;

    @Transactional
    public ConfigItem setConfig(Long projectId, String environment, ConfigItemDTO dto) {
        Project project = projectService.getProject(projectId);
        
        ConfigItem configItem = configItemRepository
                .findByProjectIdAndEnvironmentAndConfigKey(projectId, environment, dto.getConfigKey())
                .orElseGet(() -> {
                    ConfigItem newConfig = new ConfigItem();
                    newConfig.setProject(project);
                    newConfig.setEnvironment(environment);
                    newConfig.setConfigKey(dto.getConfigKey());
                    newConfig.setStatus(ConfigItem.ReleaseStatus.DRAFT);
                    return newConfig;
                });

        if (configItem.getStatus() == ConfigItem.ReleaseStatus.PENDING || 
            configItem.getStatus() == ConfigItem.ReleaseStatus.GRAY_RELEASE) {
            throw new RuntimeException("Config '" + dto.getConfigKey() + "' is in " + configItem.getStatus() + " state and cannot be modified");
        }

        if (configItem.getCurrentValue() != null) {
            configItem.setPendingValue(dto.getValue());
            configItem.setStatus(ConfigItem.ReleaseStatus.PENDING);
            log.info("Config '{}' for project '{}' environment '{}' moved to PENDING state with pending value", 
                    dto.getConfigKey(), project.getName(), environment);
        } else {
            configItem.setCurrentValue(dto.getValue());
            configItem.setStatus(ConfigItem.ReleaseStatus.RELEASED);
            log.info("Config '{}' for project '{}' environment '{}' created with initial value", 
                    dto.getConfigKey(), project.getName(), environment);
        }

        return configItemRepository.save(configItem);
    }

    public ConfigItem getConfig(Long projectId, String environment, String configKey) {
        return configItemRepository.findByProjectIdAndEnvironmentAndConfigKey(projectId, environment, configKey)
                .orElseThrow(() -> new RuntimeException("Config not found: " + configKey));
    }

    public List<ConfigItem> getConfigsByEnvironment(Long projectId, String environment) {
        return configItemRepository.findByProjectIdAndEnvironment(projectId, environment);
    }

    public List<ConfigItem> getPendingConfigs(Long projectId, String environment) {
        return configItemRepository.findByProjectIdAndEnvironmentAndStatus(
                projectId, 
                environment, 
                ConfigItem.ReleaseStatus.PENDING
        );
    }

    public List<ConfigItem> getReleasedConfigs(Long projectId, String environment) {
        return configItemRepository.findByProjectIdAndEnvironmentAndStatusIn(
                projectId, 
                environment, 
                List.of(ConfigItem.ReleaseStatus.RELEASED, ConfigItem.ReleaseStatus.GRAY_RELEASE)
        );
    }

    @Transactional
    public void deleteConfig(Long projectId, String environment, String configKey) {
        ConfigItem configItem = getConfig(projectId, environment, configKey);
        if (configItem.getStatus() == ConfigItem.ReleaseStatus.PENDING || 
            configItem.getStatus() == ConfigItem.ReleaseStatus.GRAY_RELEASE) {
            throw new RuntimeException("Config is in " + configItem.getStatus() + " state and cannot be deleted");
        }
        configItemRepository.delete(configItem);
    }
}
