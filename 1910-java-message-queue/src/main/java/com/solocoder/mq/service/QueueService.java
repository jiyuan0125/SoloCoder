package com.solocoder.mq.service;

import com.solocoder.mq.model.ConsumerGroup;
import com.solocoder.mq.model.Message;
import com.solocoder.mq.model.MessageStatus;
import com.solocoder.mq.model.Topic;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import javax.annotation.PreDestroy;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReadWriteLock;
import java.util.concurrent.locks.ReentrantLock;
import java.util.concurrent.locks.ReentrantReadWriteLock;

@Service
public class QueueService {

    private static final Logger log = LoggerFactory.getLogger(QueueService.class);

    @Value("${app.queue.default-topic-capacity:100000}")
    private int defaultTopicCapacity;

    @Value("${app.queue.backlog-warn-threshold:50000}")
    private int backlogWarnThreshold;

    @Value("${app.queue.ack-timeout-seconds:30}")
    private int ackTimeoutSeconds;

    @Value("${app.queue.max-retry-count:3}")
    private int maxRetryCount;

    private final PersistenceService persistenceService;

    private final ConcurrentMap<String, Topic> topics = new ConcurrentHashMap<>();
    private final ReadWriteLock topicsLock = new ReentrantReadWriteLock();

    private ScheduledExecutorService scheduler;

    public QueueService(PersistenceService persistenceService) {
        this.persistenceService = persistenceService;
    }

    @PostConstruct
    public void init() {
        restoreFromPersistence();
        scheduler = Executors.newSingleThreadScheduledExecutor();
        scheduler.scheduleWithFixedDelay(this::backgroundTasks, 1, 1, TimeUnit.SECONDS);
        log.info("Queue service initialized with default capacity: {}", defaultTopicCapacity);
    }

    @PreDestroy
    public void destroy() {
        persistenceService.saveTopics(topics);
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
        log.info("Queue service shutdown, data persisted");
    }

    private void restoreFromPersistence() {
        List<String> topicNames = persistenceService.loadAllTopicNames();
        for (String name : topicNames) {
            Topic topic = persistenceService.loadTopic(name);
            if (topic != null) {
                topics.put(name, topic);
                log.info("Restored topic: {}, messages: {}", name, topic.getMessages().size());
            }
        }
        log.info("Queue service restored {} topics from persistence", topicNames.size());
    }

    private void backgroundTasks() {
        try {
            topicsLock.readLock().lock();
            try {
                for (Topic topic : topics.values()) {
                    checkDelayedMessages(topic);
                    checkAckTimeouts(topic);
                    checkBacklog(topic);
                }
            } finally {
                topicsLock.readLock().unlock();
            }
        } catch (Exception e) {
            log.error("Background task error", e);
        }
    }

    private void checkDelayedMessages(Topic topic) {
        long now = System.currentTimeMillis();
        List<Message> messages = topic.getMessages();
        boolean modified = false;
        
        for (Message msg : messages) {
            if (msg.getStatus() == MessageStatus.PENDING && msg.getReadyAt() <= now) {
                msg.setStatus(MessageStatus.READY);
                modified = true;
                log.debug("Message {} in topic {} became ready", msg.getId(), topic.getName());
            }
        }
        
        if (modified) {
            persistenceService.saveTopic(topic);
        }
    }

    private void checkAckTimeouts(Topic topic) {
        long now = System.currentTimeMillis();
        long timeoutMs = ackTimeoutSeconds * 1000L;
        List<Message> messages = topic.getMessages();
        boolean modified = false;

        for (Message msg : messages) {
            if (msg.getStatus() == MessageStatus.CONSUMING) {
                long elapsed = now - msg.getLastConsumingAt();
                if (elapsed >= timeoutMs) {
                    handleAckTimeout(topic, msg);
                    modified = true;
                }
            }
        }

        if (modified) {
            persistenceService.saveTopic(topic);
        }
    }

    private void handleAckTimeout(Topic topic, Message msg) {
        msg.setRetryCount(msg.getRetryCount() + 1);
        log.warn("ACK timeout for message {} in topic {}, retry count: {}", 
                msg.getId(), topic.getName(), msg.getRetryCount());

        if (msg.getRetryCount() >= maxRetryCount) {
            msg.setStatus(MessageStatus.DLQ);
            log.warn("Message {} moved to DLQ in topic {} after {} retries", 
                    msg.getId(), topic.getName(), maxRetryCount);
        } else {
            msg.setStatus(MessageStatus.READY);
            log.info("Message {} requeued in topic {} for retry", msg.getId(), topic.getName());
        }
    }

    private void checkBacklog(Topic topic) {
        int backlog = getPendingMessageCount(topic);
        if (backlog > backlogWarnThreshold) {
            log.warn("Topic [{}] backlog exceeds threshold: current={}, threshold={}", 
                    topic.getName(), backlog, backlogWarnThreshold);
        }
    }

    private int getPendingMessageCount(Topic topic) {
        int count = 0;
        for (Message msg : topic.getMessages()) {
            if (msg.getStatus() == MessageStatus.READY || 
                msg.getStatus() == MessageStatus.PENDING ||
                msg.getStatus() == MessageStatus.CONSUMING) {
                count++;
            }
        }
        return count;
    }

    public Message sendMessage(String topicName, String content, long delayMs) {
        Topic topic = getOrCreateTopic(topicName);
        topicsLock.writeLock().lock();
        try {
            Message msg = Message.create(topicName, content, delayMs);
            msg.setOffset(topic.getNextOffset());
            topic.getMessages().add(msg);
            topic.setNextOffset(topic.getNextOffset() + 1);
            persistenceService.saveTopic(topic);
            log.info("Message {} sent to topic {} with delay_ms={}, offset={}", 
                    msg.getId(), topicName, delayMs, msg.getOffset());
            return msg;
        } finally {
            topicsLock.writeLock().unlock();
        }
    }

    private Topic getOrCreateTopic(String topicName) {
        topicsLock.writeLock().lock();
        try {
            Topic topic = topics.get(topicName);
            if (topic == null) {
                topic = Topic.create(topicName, defaultTopicCapacity);
                topics.put(topicName, topic);
                persistenceService.saveTopic(topic);
                log.info("Auto-created topic: {} with capacity: {}", topicName, defaultTopicCapacity);
            }
            return topic;
        } finally {
            topicsLock.writeLock().unlock();
        }
    }

    public ConsumerGroup subscribe(String topicName, String groupId, String consumerId) {
        Topic topic = getOrCreateTopic(topicName);
        ConsumerGroup group = topic.getConsumerGroups().computeIfAbsent(groupId, g -> {
            ConsumerGroup newGroup = ConsumerGroup.create(g);
            newGroup.setCurrentOffsetValue(-1);
            return newGroup;
        });
        
        group.setConsumerActive(true);
        group.setConsumerId(consumerId);
        group.setLastHeartbeat(System.currentTimeMillis());
        
        log.info("Consumer {} subscribed to topic {} in group {}", consumerId, topicName, groupId);
        persistenceService.saveTopic(topic);
        return group;
    }

    public Message poll(String topicName, String groupId, String consumerId) {
        Topic topic = topics.get(topicName);
        if (topic == null) {
            return null;
        }

        ConsumerGroup group = subscribe(topicName, groupId, consumerId);
        group.setLastHeartbeat(System.currentTimeMillis());

        if (group.getLastConsumingMessageId() != null) {
            log.debug("Consumer {} in group {} has unacked message, waiting for ACK", consumerId, groupId);
            return null;
        }

        long currentOffset = group.getCurrentOffsetValue();
        
        for (long offset = currentOffset + 1; offset < topic.getNextOffset(); offset++) {
            Message msg = getMessageByOffset(topic, offset);
            if (msg != null && msg.getStatus() == MessageStatus.READY) {
                deliverMessage(topic, group, msg);
                persistenceService.saveTopic(topic);
                return msg;
            }
        }
        
        return null;
    }

    private Message getMessageByOffset(Topic topic, long offset) {
        for (Message msg : topic.getMessages()) {
            if (msg.getOffset() == offset) {
                return msg;
            }
        }
        return null;
    }

    private void deliverMessage(Topic topic, ConsumerGroup group, Message msg) {
        msg.setStatus(MessageStatus.CONSUMING);
        msg.setLastConsumingAt(System.currentTimeMillis());
        group.setLastConsumingMessageId(msg.getId());
        group.setLastConsumingAt(System.currentTimeMillis());
        log.debug("Message {} delivered to consumer in group {}, offset={}", 
                msg.getId(), group.getGroupId(), msg.getOffset());
    }

    public boolean ack(String topicName, String groupId, String consumerId, String messageId) {
        Topic topic = topics.get(topicName);
        if (topic == null) {
            return false;
        }

        ConsumerGroup group = topic.getConsumerGroups().get(groupId);
        if (group == null) {
            return false;
        }

        if (messageId.equals(group.getLastConsumingMessageId())) {
            Message msg = getMessageById(topic, messageId);
            if (msg != null) {
                msg.setStatus(MessageStatus.ACKED);
                group.setLastConsumingMessageId(null);
                group.incrementAndGetOffset();
                group.setLastHeartbeat(System.currentTimeMillis());
                persistenceService.saveTopic(topic);
                log.debug("Message {} acked in topic {} by group {}", messageId, topicName, groupId);
                return true;
            }
        }
        
        return false;
    }

    private Message getMessageById(Topic topic, String messageId) {
        for (Message msg : topic.getMessages()) {
            if (msg.getId().equals(messageId)) {
                return msg;
            }
        }
        return null;
    }

    public List<Topic> listTopics() {
        topicsLock.readLock().lock();
        try {
            return new ArrayList<>(topics.values());
        } finally {
            topicsLock.readLock().unlock();
        }
    }

    public Topic getTopic(String name) {
        topicsLock.readLock().lock();
        try {
            return topics.get(name);
        } finally {
            topicsLock.readLock().unlock();
        }
    }

    public ConsumerGroup getConsumerGroup(String topicName, String groupId) {
        Topic topic = topics.get(topicName);
        if (topic == null) {
            return null;
        }
        return topic.getConsumerGroups().get(groupId);
    }

    public long getLag(String topicName, String groupId) {
        Topic topic = topics.get(topicName);
        if (topic == null) {
            return 0;
        }
        ConsumerGroup group = topic.getConsumerGroups().get(groupId);
        if (group == null) {
            return topic.getNextOffset();
        }
        return topic.getNextOffset() - group.getCurrentOffsetValue() - 1;
    }

    public int getTotalMessageCount(Topic topic) {
        return topic.getMessages().size();
    }

    public List<Message> getDlqMessages(String topicName) {
        Topic topic = topics.get(topicName);
        if (topic == null) {
            return Collections.emptyList();
        }
        List<Message> dlq = new ArrayList<>();
        for (Message msg : topic.getMessages()) {
            if (msg.getStatus() == MessageStatus.DLQ) {
                dlq.add(msg);
            }
        }
        return dlq;
    }

    public boolean requeueDlqMessage(String topicName, String messageId) {
        Topic topic = topics.get(topicName);
        if (topic == null) {
            return false;
        }
        Message msg = getMessageById(topic, messageId);
        if (msg == null || msg.getStatus() != MessageStatus.DLQ) {
            return false;
        }
        msg.setStatus(MessageStatus.READY);
        msg.setRetryCount(0);
        log.info("DLQ message {} requeued in topic {}", messageId, topicName);
        persistenceService.saveTopic(topic);
        return true;
    }
}
