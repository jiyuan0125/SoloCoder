package com.example.simplemq.storage;

import com.example.simplemq.model.Message;
import com.example.simplemq.model.TopicConfig;
import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.annotation.PostConstruct;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

@Slf4j
@Component
public class MessageStorage {

    @Value("${mq.storage.path}")
    private String storagePath;

    @Value("${mq.default.topic.max-messages}")
    private long defaultMaxMessages;

    @Value("${mq.default.topic.retention-ms}")
    private long defaultRetentionMs;

    private final ObjectMapper objectMapper = new ObjectMapper();
    private final Map<String, List<Message>> topicMessages = new ConcurrentHashMap<>();
    private final Map<String, AtomicLong> topicOffsets = new ConcurrentHashMap<>();
    private final Map<String, TopicConfig> topicConfigs = new ConcurrentHashMap<>();
    private final Map<String, Map<String, Long>> consumerGroups = new ConcurrentHashMap<>();

    @PostConstruct
    public void init() throws IOException {
        log.info("Initializing message storage at: {}", storagePath);
        ensureStorageDirectory();
        loadAllTopics();
        log.info("Message storage initialized with {} topics", topicConfigs.size());
    }

    private void ensureStorageDirectory() throws IOException {
        Path path = Paths.get(storagePath);
        if (!Files.exists(path)) {
            Files.createDirectories(path);
            log.info("Created storage directory: {}", storagePath);
        }
    }

    private void loadAllTopics() throws IOException {
        File storageDir = new File(storagePath);
        File[] topicDirs = storageDir.listFiles(File::isDirectory);
        
        if (topicDirs == null) {
            return;
        }

        for (File topicDir : topicDirs) {
            String topicName = topicDir.getName();
            loadTopic(topicName);
        }
    }

    private void loadTopic(String topicName) throws IOException {
        log.info("Loading topic: {}", topicName);
        
        TopicConfig config = loadTopicConfig(topicName);
        if (config == null) {
            log.warn("Topic config not found for: {}", topicName);
            return;
        }
        
        List<Message> messages = loadMessages(topicName);
        Map<String, Long> offsets = loadConsumerOffsets(topicName);

        topicConfigs.put(topicName, config);
        topicMessages.put(topicName, Collections.synchronizedList(new ArrayList<>(messages)));
        
        long maxOffset = messages.isEmpty() ? 0 : messages.get(messages.size() - 1).getOffset() + 1;
        topicOffsets.put(topicName, new AtomicLong(maxOffset));
        
        consumerGroups.put(topicName, new ConcurrentHashMap<>(offsets));
        
        log.info("Loaded topic: {} with {} messages and {} consumer groups", 
                topicName, messages.size(), offsets.size());
    }

    public TopicConfig createTopic(String topicName, Long maxMessages, Long retentionMs) throws IOException {
        if (topicConfigs.containsKey(topicName)) {
            throw new IllegalArgumentException("Topic already exists: " + topicName);
        }

        long maxMsg = maxMessages != null ? maxMessages : defaultMaxMessages;
        long retention = retentionMs != null ? retentionMs : defaultRetentionMs;
        
        TopicConfig config = new TopicConfig(topicName, maxMsg, retention);
        
        Path topicDir = Paths.get(storagePath, topicName);
        Files.createDirectories(topicDir);
        
        saveTopicConfig(config);
        
        topicConfigs.put(topicName, config);
        topicMessages.put(topicName, Collections.synchronizedList(new ArrayList<>()));
        topicOffsets.put(topicName, new AtomicLong(0));
        consumerGroups.put(topicName, new ConcurrentHashMap<>());
        
        log.info("Created topic: {}", topicName);
        return config;
    }

    public void deleteTopic(String topicName) throws IOException {
        if (!topicConfigs.containsKey(topicName)) {
            throw new IllegalArgumentException("Topic not found: " + topicName);
        }

        Path topicDir = Paths.get(storagePath, topicName);
        deleteDirectory(topicDir);
        
        topicConfigs.remove(topicName);
        topicMessages.remove(topicName);
        topicOffsets.remove(topicName);
        consumerGroups.remove(topicName);
        
        log.info("Deleted topic: {}", topicName);
    }

    public TopicConfig updateTopicConfig(String topicName, Long maxMessages, Long retentionMs) throws IOException {
        TopicConfig config = topicConfigs.get(topicName);
        if (config == null) {
            throw new IllegalArgumentException("Topic not found: " + topicName);
        }

        if (maxMessages != null) {
            config.setMaxMessages(maxMessages);
        }
        if (retentionMs != null) {
            config.setRetentionMs(retentionMs);
        }
        config.setUpdatedAt(System.currentTimeMillis());
        
        saveTopicConfig(config);
        cleanupTopic(topicName);
        
        log.info("Updated topic config: {}", topicName);
        return config;
    }

    public Message produceMessage(String topicName, String content) throws IOException {
        if (!topicConfigs.containsKey(topicName)) {
            throw new IllegalArgumentException("Topic not found: " + topicName);
        }

        cleanupTopic(topicName);
        
        Message message = new Message(topicName, content);
        message.setOffset(topicOffsets.get(topicName).getAndIncrement());
        
        topicMessages.get(topicName).add(message);
        saveMessage(topicName, message);
        
        log.debug("Produced message to topic: {}, offset: {}", topicName, message.getOffset());
        return message;
    }

    public Message consumeMessage(String topicName, String groupId) {
        if (!topicConfigs.containsKey(topicName)) {
            throw new IllegalArgumentException("Topic not found: " + topicName);
        }

        cleanupTopic(topicName);
        
        Map<String, Long> groupOffsets = consumerGroups.get(topicName);
        long currentOffset = groupOffsets.getOrDefault(groupId, 0L);
        
        List<Message> messages = topicMessages.get(topicName);
        synchronized (messages) {
            for (Message message : messages) {
                if (message.getOffset() >= currentOffset) {
                    return message;
                }
            }
        }
        
        return null;
    }

    public void acknowledgeMessage(String topicName, String groupId, long offset) throws IOException {
        if (!topicConfigs.containsKey(topicName)) {
            throw new IllegalArgumentException("Topic not found: " + topicName);
        }

        Map<String, Long> groupOffsets = consumerGroups.get(topicName);
        long currentOffset = groupOffsets.getOrDefault(groupId, 0L);
        
        if (offset + 1 > currentOffset) {
            groupOffsets.put(groupId, offset + 1);
            saveConsumerOffsets(topicName);
            log.debug("Acknowledged message: topic={}, group={}, offset={}", topicName, groupId, offset);
        }
    }

    public void resetConsumerOffset(String topicName, String groupId, long offset) throws IOException {
        if (!topicConfigs.containsKey(topicName)) {
            throw new IllegalArgumentException("Topic not found: " + topicName);
        }

        if (offset < 0) {
            throw new IllegalArgumentException("Offset cannot be negative");
        }

        Map<String, Long> groupOffsets = consumerGroups.get(topicName);
        groupOffsets.put(groupId, offset);
        saveConsumerOffsets(topicName);
        
        log.info("Reset consumer offset: topic={}, group={}, offset={}", topicName, groupId, offset);
    }

    public Set<String> listTopics() {
        return new HashSet<>(topicConfigs.keySet());
    }

    public TopicConfig getTopicConfig(String topicName) {
        return topicConfigs.get(topicName);
    }

    public long getTopicMessageCount(String topicName) {
        List<Message> messages = topicMessages.get(topicName);
        return messages != null ? messages.size() : 0;
    }

    public Map<String, Long> getConsumerOffsets(String topicName) {
        Map<String, Long> offsets = consumerGroups.get(topicName);
        return offsets != null ? new HashMap<>(offsets) : new HashMap<>();
    }

    private void cleanupTopic(String topicName) {
        TopicConfig config = topicConfigs.get(topicName);
        if (config == null) {
            return;
        }

        List<Message> messages = topicMessages.get(topicName);
        if (messages == null) {
            return;
        }

        synchronized (messages) {
            boolean removed = false;
            long now = System.currentTimeMillis();
            
            Iterator<Message> iterator = messages.iterator();
            while (iterator.hasNext()) {
                Message message = iterator.next();
                boolean expired = (now - message.getTimestamp()) > config.getRetentionMs();
                boolean overMax = messages.size() > config.getMaxMessages();
                
                if (expired || overMax) {
                    iterator.remove();
                    removed = true;
                }
            }
            
            if (removed) {
                try {
                    saveAllMessages(topicName);
                } catch (IOException e) {
                    log.error("Failed to save messages after cleanup for topic: {}", topicName, e);
                }
            }
        }
    }

    private void saveTopicConfig(TopicConfig config) throws IOException {
        Path configPath = Paths.get(storagePath, config.getName(), "config.json");
        objectMapper.writeValue(configPath.toFile(), config);
    }

    private TopicConfig loadTopicConfig(String topicName) {
        Path configPath = Paths.get(storagePath, topicName, "config.json");
        try {
            return objectMapper.readValue(configPath.toFile(), TopicConfig.class);
        } catch (IOException e) {
            log.error("Failed to load topic config for: {}", topicName, e);
            return null;
        }
    }

    private void saveMessage(String topicName, Message message) throws IOException {
        Path messagesPath = Paths.get(storagePath, topicName, "messages.json");
        List<Message> allMessages = loadMessages(topicName);
        allMessages.add(message);
        objectMapper.writeValue(messagesPath.toFile(), allMessages);
    }

    private void saveAllMessages(String topicName) throws IOException {
        Path messagesPath = Paths.get(storagePath, topicName, "messages.json");
        List<Message> messages = topicMessages.get(topicName);
        if (messages != null) {
            synchronized (messages) {
                objectMapper.writeValue(messagesPath.toFile(), new ArrayList<>(messages));
            }
        }
    }

    private List<Message> loadMessages(String topicName) {
        Path messagesPath = Paths.get(storagePath, topicName, "messages.json");
        if (!Files.exists(messagesPath)) {
            return new ArrayList<>();
        }
        try {
            return objectMapper.readValue(messagesPath.toFile(), 
                    objectMapper.getTypeFactory().constructCollectionType(List.class, Message.class));
        } catch (IOException e) {
            log.error("Failed to load messages for topic: {}", topicName, e);
            return new ArrayList<>();
        }
    }

    private void saveConsumerOffsets(String topicName) throws IOException {
        Path offsetsPath = Paths.get(storagePath, topicName, "offsets.json");
        Map<String, Long> offsets = consumerGroups.get(topicName);
        if (offsets != null) {
            objectMapper.writeValue(offsetsPath.toFile(), offsets);
        }
    }

    private Map<String, Long> loadConsumerOffsets(String topicName) {
        Path offsetsPath = Paths.get(storagePath, topicName, "offsets.json");
        if (!Files.exists(offsetsPath)) {
            return new HashMap<>();
        }
        try {
            return objectMapper.readValue(offsetsPath.toFile(),
                    objectMapper.getTypeFactory().constructMapType(Map.class, String.class, Long.class));
        } catch (IOException e) {
            log.error("Failed to load consumer offsets for topic: {}", topicName, e);
            return new HashMap<>();
        }
    }

    private void deleteDirectory(Path path) throws IOException {
        if (!Files.exists(path)) {
            return;
        }
        
        File directory = path.toFile();
        File[] files = directory.listFiles();
        
        if (files != null) {
            for (File file : files) {
                if (file.isDirectory()) {
                    deleteDirectory(file.toPath());
                } else {
                    file.delete();
                }
            }
        }
        
        Files.delete(path);
    }
}
