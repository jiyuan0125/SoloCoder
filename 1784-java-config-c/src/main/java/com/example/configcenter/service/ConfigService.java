package com.example.configcenter.service;

import com.example.configcenter.dto.BatchImportResponse;
import com.example.configcenter.dto.ConfigNotification;
import com.example.configcenter.dto.ConfigRequest;
import com.example.configcenter.dto.ConfigResponse;
import com.example.configcenter.entity.ConfigHistory;
import com.example.configcenter.entity.ConfigItem;
import com.example.configcenter.entity.ConfigSubscriber;
import com.example.configcenter.exception.ValidationException;
import com.example.configcenter.repository.ConfigHistoryRepository;
import com.example.configcenter.repository.ConfigItemRepository;
import com.example.configcenter.repository.ConfigSubscriberRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.client.RestTemplate;

import java.nio.charset.StandardCharsets;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeParseException;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.regex.Pattern;

@Slf4j
@Service
@RequiredArgsConstructor
public class ConfigService {
    
    private static final long MAX_VALUE_SIZE_BYTES = 64 * 1024; // 64KB
    private static final int BATCH_LIMIT = 100;
    private static final Pattern KEY_PATTERN = Pattern.compile("^[a-zA-Z0-9_.-]+$");
    private static final DateTimeFormatter DATE_TIME_FORMATTER = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
    private static final String EXPECTED_FORMAT = "yyyy-MM-dd HH:mm:ss";
    
    private final ConfigItemRepository configItemRepository;
    private final ConfigHistoryRepository configHistoryRepository;
    private final ConfigSubscriberRepository configSubscriberRepository;
    private final RestTemplate restTemplate;
    
    public ConfigResponse saveConfig(ConfigRequest request) {
        validateKey(request.getKey());
        
        String key = request.getKey();
        String value = request.getValue();
        String warningMessage = null;
        
        if (value != null && value.getBytes(StandardCharsets.UTF_8).length > MAX_VALUE_SIZE_BYTES) {
            warningMessage = "配置值超过 64KB，可能影响下发效率，建议拆分";
            log.warn("Config value for key {} exceeds 64KB, size: {} bytes", 
                     key, value.getBytes(StandardCharsets.UTF_8).length);
        }
        
        Optional<ConfigItem> existing = configItemRepository.findByConfigKey(key);
        int newVersion;
        ConfigItem configItem;
        
        if (existing.isPresent()) {
            configItem = existing.get();
            saveToHistory(configItem);
            newVersion = configItem.getVersion() + 1;
            configItem.setVersion(newVersion);
            configItem.setConfigValue(value);
        } else {
            configItem = new ConfigItem();
            configItem.setConfigKey(key);
            configItem.setConfigValue(value);
            configItem.setVersion(1);
            newVersion = 1;
        }
        
        configItemRepository.save(configItem);
        notifySubscribers(key, newVersion, "UPDATE");
        
        return ConfigResponse.builder()
            .key(key)
            .value(value)
            .version(newVersion)
            .message(warningMessage)
            .build();
    }
    
    public Optional<ConfigResponse> getConfig(String key) {
        return configItemRepository.findByConfigKey(key)
            .map(item -> ConfigResponse.builder()
                .key(item.getConfigKey())
                .value(item.getConfigValue())
                .version(item.getVersion())
                .build());
    }
    
    @Transactional
    public boolean deleteConfig(String key) {
        Optional<ConfigItem> existing = configItemRepository.findByConfigKey(key);
        if (existing.isEmpty()) {
            return false;
        }
        
        ConfigItem item = existing.get();
        int version = item.getVersion();
        
        saveToHistory(item);
        configItemRepository.deleteByConfigKey(key);
        
        notifySubscribers(key, version, "DELETE");
        configSubscriberRepository.deleteByConfigKey(key);
        
        return true;
    }
    
    @Transactional
    public BatchImportResponse batchImport(List<ConfigRequest> requests) {
        if (requests == null || requests.isEmpty()) {
            return BatchImportResponse.builder()
                .total(0)
                .successCount(0)
                .failureCount(0)
                .failures(List.of())
                .build();
        }
        
        if (requests.size() > BATCH_LIMIT) {
            throw new ValidationException("批量导入超过上限 " + BATCH_LIMIT + " 条");
        }
        
        List<BatchImportResponse.FailureItem> failures = new ArrayList<>();
        int successCount = 0;
        
        for (int i = 0; i < requests.size(); i++) {
            ConfigRequest request = requests.get(i);
            try {
                validateKey(request.getKey());
                saveConfig(request);
                successCount++;
            } catch (Exception e) {
                failures.add(BatchImportResponse.FailureItem.builder()
                    .key(request.getKey())
                    .reason("第 " + (i + 1) + " 条: " + e.getMessage())
                    .build());
            }
        }
        
        return BatchImportResponse.builder()
            .total(requests.size())
            .successCount(successCount)
            .failureCount(failures.size())
            .failures(failures)
            .build();
    }
    
    public List<ConfigHistory> getHistory(String key, String startTimeStr, String endTimeStr) {
        if (startTimeStr == null && endTimeStr == null) {
            return configHistoryRepository.findTop50ByConfigKeyOrderByCreatedAtDesc(key);
        }
        
        LocalDateTime startTime = parseDateTime(startTimeStr, "开始时间");
        LocalDateTime endTime = parseDateTime(endTimeStr, "结束时间");
        
        if (startTime == null) {
            startTime = LocalDateTime.of(1970, 1, 1, 0, 0, 0);
        }
        if (endTime == null) {
            endTime = LocalDateTime.now();
        }
        
        return configHistoryRepository.findByConfigKeyAndCreatedAtBetweenOrderByCreatedAtDesc(
            key, startTime, endTime);
    }
    
    public void subscribe(String key, String callbackUrl) {
        ConfigSubscriber subscriber = new ConfigSubscriber();
        subscriber.setConfigKey(key);
        subscriber.setSubscriberUrl(callbackUrl);
        configSubscriberRepository.save(subscriber);
        log.info("New subscriber for key {}: {}", key, callbackUrl);
    }
    
    private void validateKey(String key) {
        if (key == null || key.isEmpty()) {
            throw new ValidationException("配置 key 不能为空");
        }
        if (!KEY_PATTERN.matcher(key).matches()) {
            throw new ValidationException("配置 key 包含非法字符，只能包含字母、数字、下划线、点号和短横线，不能包含空格、换行等特殊字符");
        }
    }
    
    private void saveToHistory(ConfigItem item) {
        ConfigHistory history = new ConfigHistory();
        history.setConfigKey(item.getConfigKey());
        history.setConfigValue(item.getConfigValue());
        history.setVersion(item.getVersion());
        configHistoryRepository.save(history);
    }
    
    private void notifySubscribers(String key, int version, String action) {
        List<ConfigSubscriber> subscribers = configSubscriberRepository.findByConfigKey(key);
        log.info("Notifying {} subscribers for key {} action {} version {}", 
                 subscribers.size(), key, action, version);
        
        ConfigNotification notification = ConfigNotification.builder()
            .action(action)
            .key(key)
            .version(version)
            .build();
        
        for (ConfigSubscriber subscriber : subscribers) {
            try {
                restTemplate.postForObject(subscriber.getSubscriberUrl(), notification, Void.class);
                log.info("Successfully notified subscriber: {}", subscriber.getSubscriberUrl());
            } catch (Exception e) {
                log.error("Failed to notify subscriber {}: {}", subscriber.getSubscriberUrl(), e.getMessage());
            }
        }
    }
    
    private LocalDateTime parseDateTime(String dateTimeStr, String fieldName) {
        if (dateTimeStr == null) {
            return null;
        }
        try {
            return LocalDateTime.parse(dateTimeStr, DATE_TIME_FORMATTER);
        } catch (DateTimeParseException e) {
            throw new ValidationException(fieldName + " 格式错误，期望格式 " + EXPECTED_FORMAT);
        }
    }
}
