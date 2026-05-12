package com.configcenter.service;

import com.configcenter.exception.ConfigNotFoundException;
import com.configcenter.exception.InvalidRequestException;
import com.configcenter.model.AuditLog;
import com.configcenter.model.ConfigHistory;
import com.configcenter.model.ConfigItem;
import com.configcenter.repository.AuditLogRepository;
import com.configcenter.repository.ConfigHistoryRepository;
import com.configcenter.repository.ConfigItemRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Optional;

@Service
@RequiredArgsConstructor
public class ConfigService {
    
    private final ConfigItemRepository configItemRepository;
    private final ConfigHistoryRepository configHistoryRepository;
    private final AuditLogRepository auditLogRepository;
    private final ConfigChangeNotifier configChangeNotifier;
    
    public void validateKeyParams(String namespace, String group, String key) {
        if (namespace == null || namespace.trim().isEmpty()) {
            throw new InvalidRequestException("namespace cannot be empty");
        }
        if (group == null || group.trim().isEmpty()) {
            throw new InvalidRequestException("group cannot be empty");
        }
        if (key == null || key.trim().isEmpty()) {
            throw new InvalidRequestException("key cannot be empty");
        }
    }
    
    public Optional<ConfigItem> getConfig(String namespace, String group, String key) {
        validateKeyParams(namespace, group, key);
        return configItemRepository.findByNamespaceAndGroupAndKey(namespace, group, key);
    }
    
    @Transactional
    public ConfigItem createConfig(String namespace, String group, String key, String value, String operator, String clientIp) {
        validateKeyParams(namespace, group, key);
        
        if (configItemRepository.existsByNamespaceAndGroupAndKey(namespace, group, key)) {
            throw new InvalidRequestException("Config already exists");
        }
        
        ConfigItem configItem = ConfigItem.builder()
                .namespace(namespace)
                .group(group)
                .key(key)
                .value(value)
                .version(1L)
                .build();
        
        configItem = configItemRepository.save(configItem);
        
        saveHistory(configItem, operator, clientIp);
        saveAuditLog("CREATE", namespace, group, key, null, value, null, 1L, operator, clientIp);
        
        configChangeNotifier.notifyChange(namespace, group, key);
        
        return configItem;
    }
    
    @Transactional
    public ConfigItem updateConfig(String namespace, String group, String key, String value, String operator, String clientIp) {
        validateKeyParams(namespace, group, key);
        
        ConfigItem configItem = configItemRepository.findByNamespaceAndGroupAndKey(namespace, group, key)
                .orElseThrow(() -> new ConfigNotFoundException("Config not found"));
        
        if (value.equals(configItem.getValue())) {
            return configItem;
        }
        
        String oldValue = configItem.getValue();
        Long oldVersion = configItem.getVersion();
        Long newVersion = oldVersion + 1;
        
        configItem.setValue(value);
        configItem.setVersion(newVersion);
        configItem = configItemRepository.save(configItem);
        
        saveHistory(configItem, operator, clientIp);
        saveAuditLog("UPDATE", namespace, group, key, oldValue, value, oldVersion, newVersion, operator, clientIp);
        
        configChangeNotifier.notifyChange(namespace, group, key);
        
        return configItem;
    }
    
    @Transactional
    public void deleteConfig(String namespace, String group, String key, String operator, String clientIp) {
        validateKeyParams(namespace, group, key);
        
        ConfigItem configItem = configItemRepository.findByNamespaceAndGroupAndKey(namespace, group, key)
                .orElseThrow(() -> new ConfigNotFoundException("Config not found"));
        
        String oldValue = configItem.getValue();
        Long oldVersion = configItem.getVersion();
        
        saveAuditLog("DELETE", namespace, group, key, oldValue, null, oldVersion, null, operator, clientIp);
        
        configItemRepository.delete(configItem);
        
        configChangeNotifier.notifyChange(namespace, group, key);
    }
    
    @Transactional
    public ConfigItem rollbackConfig(String namespace, String group, String key, Long targetVersion, String operator, String clientIp) {
        validateKeyParams(namespace, group, key);
        
        ConfigHistory targetHistory = configHistoryRepository
                .findByNamespaceAndGroupAndKeyAndVersion(namespace, group, key, targetVersion)
                .orElseThrow(() -> new ConfigNotFoundException("History version not found"));
        
        ConfigItem currentConfig = configItemRepository.findByNamespaceAndGroupAndKey(namespace, group, key)
                .orElseThrow(() -> new ConfigNotFoundException("Config not found"));
        
        String oldValue = currentConfig.getValue();
        Long oldVersion = currentConfig.getVersion();
        Long newVersion = oldVersion + 1;
        String newValue = targetHistory.getValue();
        
        currentConfig.setValue(newValue);
        currentConfig.setVersion(newVersion);
        currentConfig = configItemRepository.save(currentConfig);
        
        saveHistory(currentConfig, operator, clientIp);
        
        AuditLog auditLog = AuditLog.builder()
                .namespace(namespace)
                .group(group)
                .key(key)
                .action("ROLLBACK")
                .oldValue(oldValue)
                .newValue(newValue)
                .targetVersion(targetVersion)
                .oldVersion(oldVersion)
                .newVersion(newVersion)
                .operator(operator)
                .clientIp(clientIp)
                .build();
        auditLogRepository.save(auditLog);
        
        configChangeNotifier.notifyChange(namespace, group, key);
        
        return currentConfig;
    }
    
    public Page<ConfigItem> listConfigs(String namespace, String group, String key, Pageable pageable) {
        Specification<ConfigItem> spec = Specification.where(null);
        
        if (namespace != null && !namespace.trim().isEmpty()) {
            spec = spec.and((root, query, cb) -> cb.equal(root.get("namespace"), namespace));
        }
        if (group != null && !group.trim().isEmpty()) {
            spec = spec.and((root, query, cb) -> cb.equal(root.get("group"), group));
        }
        if (key != null && !key.trim().isEmpty()) {
            spec = spec.and((root, query, cb) -> cb.like(root.get("key"), "%" + key + "%"));
        }
        
        return configItemRepository.findAll(spec, pageable);
    }
    
    public Page<ConfigHistory> listHistory(String namespace, String group, String key, Pageable pageable) {
        validateKeyParams(namespace, group, key);
        return configHistoryRepository.findByNamespaceAndGroupAndKeyOrderByVersionDesc(namespace, group, key, pageable);
    }
    
    public Optional<ConfigHistory> getHistoryVersion(String namespace, String group, String key, Long version) {
        validateKeyParams(namespace, group, key);
        return configHistoryRepository.findByNamespaceAndGroupAndKeyAndVersion(namespace, group, key, version);
    }
    
    public Page<AuditLog> listAuditLogs(String namespace, String group, String key, Pageable pageable) {
        validateKeyParams(namespace, group, key);
        return auditLogRepository.findByNamespaceAndGroupAndKeyOrderByCreatedAtDesc(namespace, group, key, pageable);
    }
    
    private void saveHistory(ConfigItem configItem, String operator, String clientIp) {
        ConfigHistory history = ConfigHistory.builder()
                .namespace(configItem.getNamespace())
                .group(configItem.getGroup())
                .key(configItem.getKey())
                .value(configItem.getValue())
                .version(configItem.getVersion())
                .createdBy(operator)
                .clientIp(clientIp)
                .build();
        configHistoryRepository.save(history);
    }
    
    private void saveAuditLog(String action, String namespace, String group, String key,
                              String oldValue, String newValue, Long oldVersion, Long newVersion,
                              String operator, String clientIp) {
        AuditLog auditLog = AuditLog.builder()
                .namespace(namespace)
                .group(group)
                .key(key)
                .action(action)
                .oldValue(oldValue)
                .newValue(newValue)
                .oldVersion(oldVersion)
                .newVersion(newVersion)
                .operator(operator)
                .clientIp(clientIp)
                .build();
        auditLogRepository.save(auditLog);
    }
}
