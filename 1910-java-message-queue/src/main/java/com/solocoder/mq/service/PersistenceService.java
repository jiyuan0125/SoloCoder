package com.solocoder.mq.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.solocoder.mq.model.ConsumerGroup;
import com.solocoder.mq.model.Message;
import com.solocoder.mq.model.Topic;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import javax.annotation.PreDestroy;
import java.io.*;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.ReadWriteLock;
import java.util.concurrent.locks.ReentrantReadWriteLock;

@Service
public class PersistenceService {

    private static final Logger log = LoggerFactory.getLogger(PersistenceService.class);

    @Value("${app.queue.data-dir:./data}")
    private String dataDir;

    private final ObjectMapper objectMapper;
    private final ReadWriteLock lock = new ReentrantReadWriteLock();
    private ScheduledExecutorService scheduler;

    public PersistenceService() {
        this.objectMapper = new ObjectMapper();
        this.objectMapper.enable(SerializationFeature.INDENT_OUTPUT);
    }

    @PostConstruct
    public void init() throws IOException {
        Path dataPath = Paths.get(dataDir);
        if (!Files.exists(dataPath)) {
            Files.createDirectories(dataPath);
        }
        scheduler = Executors.newSingleThreadScheduledExecutor();
        scheduler.scheduleWithFixedDelay(this::autoSave, 5, 5, TimeUnit.SECONDS);
        log.info("Persistence service initialized, data dir: {}", dataDir);
    }

    @PreDestroy
    public void destroy() {
        if (scheduler != null) {
            scheduler.shutdown();
            try {
                if (!scheduler.awaitTermination(5, TimeUnit.SECONDS)) {
                    scheduler.shutdownNow();
                }
            } catch (InterruptedException e) {
                scheduler.shutdownNow();
            }
        }
    }

    private void autoSave() {
    }

    public void saveTopics(ConcurrentMap<String, Topic> topics) {
        lock.writeLock().lock();
        try {
            for (Topic topic : topics.values()) {
                saveTopic(topic);
            }
            writeTopicIndex(topics);
        } finally {
            lock.writeLock().unlock();
        }
    }

    public void saveTopic(Topic topic) {
        lock.writeLock().lock();
        try {
            writeTopicMetadata(topic);
            writeTopicMessages(topic);
            writeConsumerGroups(topic);
        } finally {
            lock.writeLock().unlock();
        }
    }

    public Topic loadTopic(String name) {
        lock.readLock().lock();
        try {
            Topic topic = readTopicMetadata(name);
            if (topic == null) {
                return null;
            }
            topic.setMessages(readTopicMessages(name));
            topic.setConsumerGroups(readConsumerGroups(name));
            return topic;
        } finally {
            lock.readLock().unlock();
        }
    }

    public List<String> loadAllTopicNames() {
        lock.readLock().lock();
        try {
            return readTopicIndex();
        } finally {
            lock.readLock().unlock();
        }
    }

    private Path getTopicDir(String topicName) {
        return Paths.get(dataDir, "topics", topicName);
    }

    private Path getTopicMetadataPath(String topicName) {
        return getTopicDir(topicName).resolve("metadata.json");
    }

    private Path getTopicMessagesPath(String topicName) {
        return getTopicDir(topicName).resolve("messages.jsonl");
    }

    private Path getConsumerGroupsPath(String topicName) {
        return getTopicDir(topicName).resolve("consumers.json");
    }

    private Path getTopicIndexPath() {
        return Paths.get(dataDir, "topic_index.txt");
    }

    private void writeTopicMetadata(Topic topic) {
        try {
            Files.createDirectories(getTopicDir(topic.getName()));
            String json = objectMapper.writeValueAsString(topic);
            Files.writeString(getTopicMetadataPath(topic.getName()), json, StandardCharsets.UTF_8);
        } catch (IOException e) {
            log.error("Failed to write topic metadata: {}", topic.getName(), e);
            throw new RuntimeException("Failed to persist topic metadata", e);
        }
    }

    private Topic readTopicMetadata(String name) {
        Path path = getTopicMetadataPath(name);
        if (!Files.exists(path)) {
            return null;
        }
        try {
            String json = Files.readString(path, StandardCharsets.UTF_8);
            return objectMapper.readValue(json, Topic.class);
        } catch (IOException e) {
            log.error("Failed to read topic metadata: {}", name, e);
            return null;
        }
    }

    private void writeTopicMessages(Topic topic) {
        Path path = getTopicMessagesPath(topic.getName());
        try {
            Files.createDirectories(getTopicDir(topic.getName()));
            try (BufferedWriter writer = Files.newBufferedWriter(path, StandardCharsets.UTF_8)) {
                for (Message msg : topic.getMessages()) {
                    String json = objectMapper.writeValueAsString(msg);
                    writer.write(json);
                    writer.newLine();
                }
            }
        } catch (IOException e) {
            log.error("Failed to write messages for topic: {}", topic.getName(), e);
            throw new RuntimeException("Failed to persist messages", e);
        }
    }

    private List<Message> readTopicMessages(String name) {
        List<Message> messages = new ArrayList<>();
        Path path = getTopicMessagesPath(name);
        if (!Files.exists(path)) {
            return messages;
        }
        try (BufferedReader reader = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            String line;
            while ((line = reader.readLine()) != null) {
                if (!line.trim().isEmpty()) {
                    Message msg = objectMapper.readValue(line, Message.class);
                    messages.add(msg);
                }
            }
        } catch (IOException e) {
            log.error("Failed to read messages for topic: {}", name, e);
        }
        return messages;
    }

    private void writeConsumerGroups(Topic topic) {
        try {
            Files.createDirectories(getTopicDir(topic.getName()));
            String json = objectMapper.writeValueAsString(topic.getConsumerGroups());
            Files.writeString(getConsumerGroupsPath(topic.getName()), json, StandardCharsets.UTF_8);
        } catch (IOException e) {
            log.error("Failed to write consumer groups for topic: {}", topic.getName(), e);
            throw new RuntimeException("Failed to persist consumer groups", e);
        }
    }

    private ConcurrentMap<String, ConsumerGroup> readConsumerGroups(String name) {
        ConcurrentMap<String, ConsumerGroup> groups = new java.util.concurrent.ConcurrentHashMap<>();
        Path path = getConsumerGroupsPath(name);
        if (!Files.exists(path)) {
            return groups;
        }
        try {
            String json = Files.readString(path, StandardCharsets.UTF_8);
            if (json != null && !json.trim().isEmpty()) {
                java.util.Map<String, ConsumerGroup> map = objectMapper.readValue(json,
                        new com.fasterxml.jackson.core.type.TypeReference<java.util.Map<String, ConsumerGroup>>() {});
                groups.putAll(map);
            }
        } catch (IOException e) {
            log.error("Failed to read consumer groups for topic: {}", name, e);
        }
        return groups;
    }

    private void writeTopicIndex(ConcurrentMap<String, Topic> topics) {
        try {
            Path path = getTopicIndexPath();
            Files.createDirectories(path.getParent());
            try (BufferedWriter writer = Files.newBufferedWriter(path, StandardCharsets.UTF_8)) {
                for (String name : topics.keySet()) {
                    writer.write(name);
                    writer.newLine();
                }
            }
        } catch (IOException e) {
            log.error("Failed to write topic index", e);
            throw new RuntimeException("Failed to persist topic index", e);
        }
    }

    private List<String> readTopicIndex() {
        List<String> names = new ArrayList<>();
        Path path = getTopicIndexPath();
        if (!Files.exists(path)) {
            return names;
        }
        try (BufferedReader reader = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            String line;
            while ((line = reader.readLine()) != null) {
                if (!line.trim().isEmpty()) {
                    names.add(line.trim());
                }
            }
        } catch (IOException e) {
            log.error("Failed to read topic index", e);
        }
        return names;
    }
}
