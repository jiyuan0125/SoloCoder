package com.example.lightweightqueue.service;

import com.example.lightweightqueue.model.ConsumerGroup;
import com.example.lightweightqueue.model.Topic;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;

@Slf4j
@Service
public class PersistenceService {

    private static final String PROGRESS_FILE = "consumer-progress.json";

    @Value("${queue.data-dir:./queue-data}")
    private String dataDir = "./queue-data";

    private final ObjectMapper objectMapper;

    public PersistenceService(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    public void init() {
        if (dataDir == null) {
            dataDir = "./queue-data";
        }
        try {
            Path dir = Paths.get(dataDir);
            if (!Files.exists(dir)) {
                Files.createDirectories(dir);
            }
        } catch (IOException e) {
            log.error("Failed to create data directory: {}", dataDir, e);
        }
    }

    public void saveConsumerProgress(String topic, String groupId, long consumedCount, long lastConsumedIndex) {
        try {
            init();
            Path progressFile = Paths.get(dataDir, PROGRESS_FILE);
            
            ConcurrentMap<String, ProgressData> allProgress;
            if (Files.exists(progressFile)) {
                allProgress = readProgress(progressFile);
            } else {
                allProgress = new ConcurrentHashMap<>();
            }
            
            String key = topic + ":" + groupId;
            allProgress.put(key, new ProgressData(topic, groupId, consumedCount, lastConsumedIndex));
            
            writeProgress(progressFile, allProgress);
        } catch (IOException e) {
            log.error("Failed to save consumer progress", e);
        }
    }

    public void restoreProgress(TopicManager topicManager) {
        try {
            Path progressFile = Paths.get(dataDir, PROGRESS_FILE);
            if (!Files.exists(progressFile)) {
                return;
            }
            
            ConcurrentMap<String, ProgressData> allProgress = readProgress(progressFile);
            
            allProgress.forEach((key, data) -> {
                Topic topic = topicManager.getOrCreateTopic(data.getTopic());
                ConsumerGroup group = topicManager.getOrCreateConsumerGroup(topic, data.getGroupId());
                group.setConsumedCount(data.getConsumedCount());
                group.setLastConsumedIndex(data.getLastConsumedIndex());
                log.info("Restored progress for {}: consumedCount={}, lastIndex={}", 
                        key, data.getConsumedCount(), data.getLastConsumedIndex());
            });
        } catch (IOException e) {
            log.error("Failed to restore consumer progress", e);
        }
    }

    @SuppressWarnings("unchecked")
    private ConcurrentMap<String, ProgressData> readProgress(Path file) throws IOException {
        String content = Files.readString(file);
        if (content.isBlank()) {
            return new ConcurrentHashMap<>();
        }
        return objectMapper.readValue(content, ConcurrentHashMap.class);
    }

    private void writeProgress(Path file, ConcurrentMap<String, ProgressData> data) throws IOException {
        String json = objectMapper.writeValueAsString(data);
        Files.writeString(file, json);
    }

    @lombok.Data
    @lombok.NoArgsConstructor
    @lombok.AllArgsConstructor
    public static class ProgressData {
        private String topic;
        private String groupId;
        private long consumedCount;
        private long lastConsumedIndex;
    }
}
