package com.example.memorymq.service;

import com.example.memorymq.config.MqProperties;
import com.example.memorymq.engine.TopicPartition;
import com.example.memorymq.model.Message;
import com.example.memorymq.model.TopicInfo;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.locks.ReadWriteLock;
import java.util.concurrent.locks.ReentrantReadWriteLock;

@Slf4j
@Service
public class MessageEngine {
    private final MqProperties properties;
    private final ConcurrentHashMap<String, TopicPartition> topics;
    private final ConcurrentHashMap<String, Instant> topicCreationTimes;
    private final ReadWriteLock topicLock;

    public MessageEngine(MqProperties properties) {
        this.properties = properties;
        this.topics = new ConcurrentHashMap<>();
        this.topicCreationTimes = new ConcurrentHashMap<>();
        this.topicLock = new ReentrantReadWriteLock();
    }

    public boolean createTopic(String name, Integer maxMessages) {
        topicLock.writeLock().lock();
        try {
            if (topics.containsKey(name)) {
                return false;
            }
            int capacity = (maxMessages != null && maxMessages > 0)
                ? maxMessages
                : properties.getDefaultMaxMessagesPerTopic();
            TopicPartition partition = new TopicPartition(name, capacity, properties);
            topics.put(name, partition);
            topicCreationTimes.put(name, Instant.now());
            log.info("Topic created: {} with capacity {}", name, capacity);
            return true;
        } finally {
            topicLock.writeLock().unlock();
        }
    }

    public boolean deleteTopic(String name) {
        topicLock.writeLock().lock();
        try {
            TopicPartition partition = topics.remove(name);
            if (partition != null) {
                partition.shutdown();
                topicCreationTimes.remove(name);
                log.info("Topic deleted: {}", name);
                return true;
            }
            return false;
        } finally {
            topicLock.writeLock().unlock();
        }
    }

    public List<TopicInfo> listTopics() {
        topicLock.readLock().lock();
        try {
            List<TopicInfo> list = new ArrayList<>();
            for (Map.Entry<String, TopicPartition> entry : topics.entrySet()) {
                String name = entry.getKey();
                TopicPartition partition = entry.getValue();
                list.add(TopicInfo.builder()
                    .name(name)
                    .maxMessages(getTopicCapacity(name))
                    .currentMessageCount(partition.getCurrentMessageCount())
                    .consumerGroups(partition.getConsumerGroups())
                    .createdAt(topicCreationTimes.get(name))
                    .build());
            }
            return list;
        } finally {
            topicLock.readLock().unlock();
        }
    }

    public boolean topicExists(String name) {
        topicLock.readLock().lock();
        try {
            return topics.containsKey(name);
        } finally {
            topicLock.readLock().unlock();
        }
    }

    private int getTopicCapacity(String name) {
        return properties.getDefaultMaxMessagesPerTopic();
    }

    public boolean publish(String topic, Message message) {
        topicLock.readLock().lock();
        try {
            TopicPartition partition = topics.get(topic);
            if (partition == null) {
                return false;
            }
            return partition.publish(message);
        } finally {
            topicLock.readLock().unlock();
        }
    }

    public List<Message> consume(String topic, String groupId, int maxMessages, int timeoutSeconds) {
        topicLock.readLock().lock();
        try {
            TopicPartition partition = topics.get(topic);
            if (partition == null) {
                return Collections.emptyList();
            }
            return partition.consume(groupId, maxMessages, timeoutSeconds * 1000L);
        } finally {
            topicLock.readLock().unlock();
        }
    }

    public boolean ack(String topic, String groupId, List<String> messageIds) {
        topicLock.readLock().lock();
        try {
            TopicPartition partition = topics.get(topic);
            if (partition == null) {
                return false;
            }
            return partition.ack(groupId, messageIds);
        } finally {
            topicLock.readLock().unlock();
        }
    }

    public boolean nack(String topic, String groupId, List<String> messageIds, boolean requeue) {
        topicLock.readLock().lock();
        try {
            TopicPartition partition = topics.get(topic);
            if (partition == null) {
                return false;
            }
            return partition.nack(groupId, messageIds, requeue);
        } finally {
            topicLock.readLock().unlock();
        }
    }

    @Scheduled(fixedDelayString = "#{mqProperties.backlogCheckIntervalSeconds * 1000}")
    public void cleanup() {
        topicLock.readLock().lock();
        try {
            for (TopicPartition partition : topics.values()) {
                partition.cleanupExpired();
            }
        } finally {
            topicLock.readLock().unlock();
        }
    }
}
