package com.example.lightweightqueue.service;

import com.example.lightweightqueue.model.Consumer;
import com.example.lightweightqueue.model.ConsumerGroup;
import com.example.lightweightqueue.model.DeliveryFailureType;
import com.example.lightweightqueue.model.Message;
import com.example.lightweightqueue.model.MessageFormat;
import com.example.lightweightqueue.model.MessageStatus;
import com.example.lightweightqueue.model.Topic;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.List;
import java.util.Optional;
import java.util.UUID;

@Slf4j
@Service
@RequiredArgsConstructor
public class MessageQueueService {

    private static final long ACK_TIMEOUT_SECONDS = 30;
    private static final int MAX_RETRY_COUNT = 3;
    private static final long[] RETRY_INTERVALS = {5, 15, 45};

    private final TopicManager topicManager;
    private final ObjectMapper objectMapper;
    private final PersistenceService persistenceService;

    public String produce(String topicName, String body, String formatStr) {
        Topic topic = topicManager.getOrCreateTopic(topicName);

        MessageFormat format;
        try {
            format = parseFormat(formatStr);
            validateFormat(body, format);
        } catch (IllegalArgumentException e) {
            throw new IllegalArgumentException("Format error: " + e.getMessage(), e);
        }

        Message message = Message.builder()
                .id(UUID.randomUUID().toString())
                .topic(topicName)
                .body(body)
                .format(format)
                .retryCount(0)
                .status(MessageStatus.PENDING_DELIVERY)
                .createdAt(Instant.now())
                .build();

        topic.addMessage(message);
        return message.getId();
    }

    public Optional<Message> consume(String topicName, String groupId, String consumerId) {
        Topic topic = topicManager.getTopic(topicName);
        if (topic == null) {
            return Optional.empty();
        }

        ConsumerGroup group = topicManager.getOrCreateConsumerGroup(topic, groupId);

        registerConsumerHeartbeat(group, consumerId);

        List<Consumer> consumers = group.getConsumers();
        if (consumers.isEmpty()) {
            return Optional.empty();
        }

        Message message = findNextMessageForConsumer(topic, group, consumerId);
        if (message != null) {
            markAsDelivered(message, consumerId);
            group.getInFlightMessages().put(message.getId(), message);
            return Optional.of(message);
        }

        return Optional.empty();
    }

    public boolean acknowledge(String topicName, String groupId, String consumerId, String messageId) {
        Topic topic = topicManager.getTopic(topicName);
        if (topic == null) {
            return false;
        }

        ConsumerGroup group = topicManager.getConsumerGroup(topic, groupId);
        if (group == null) {
            return false;
        }

        Message message = group.getInFlightMessages().remove(messageId);
        if (message == null) {
            return false;
        }

        message.setStatus(MessageStatus.ACKNOWLEDGED);
        group.setConsumedCount(group.getConsumedCount() + 1);

        persistenceService.saveConsumerProgress(topicName, groupId, group.getConsumedCount(), group.getLastConsumedIndex());

        return true;
    }

    public void handleTimeoutOrFailure(Topic topic, ConsumerGroup group, Message message, DeliveryFailureType failureType) {
        message.setLastError(failureType.name());

        if (message.getRetryCount() >= MAX_RETRY_COUNT) {
            moveToDeadLetter(topic, message);
            return;
        }

        long retryInterval = RETRY_INTERVALS[Math.min(message.getRetryCount(), RETRY_INTERVALS.length - 1)];
        message.setRetryCount(message.getRetryCount() + 1);
        message.setStatus(MessageStatus.PENDING_DELIVERY);
        message.setNextRetryAt(Instant.now().plusSeconds(retryInterval));
        message.setDeliveredToConsumerId(null);
        message.setDeliveredAt(null);

        group.getInFlightMessages().remove(message.getId());
    }

    private Message findNextMessageForConsumer(Topic topic, ConsumerGroup group, String consumerId) {
        List<Consumer> consumers = group.getConsumers();
        if (consumers.isEmpty()) {
            return null;
        }

        List<Message> messages = topic.getMessages();
        Instant now = Instant.now();

        synchronized (messages) {
            for (int i = 0; i < messages.size(); i++) {
                Message msg = messages.get(i);

                if (msg.getStatus() == MessageStatus.DEAD_LETTER || msg.getStatus() == MessageStatus.ACKNOWLEDGED) {
                    continue;
                }

                if (msg.getStatus() == MessageStatus.DELIVERED) {
                    continue;
                }

                if (msg.getNextRetryAt() != null && msg.getNextRetryAt().isAfter(now)) {
                    continue;
                }

                if (group.getInFlightMessages().containsKey(msg.getId())) {
                    continue;
                }

                int assignedIndex = assignConsumerIndex(group, msg);
                Consumer assignedConsumer = consumers.get(assignedIndex);

                if (assignedConsumer.getId().equals(consumerId)) {
                    group.setLastConsumedIndex(i);
                    return msg;
                }
            }
        }

        return null;
    }

    private int assignConsumerIndex(ConsumerGroup group, Message message) {
        List<Consumer> consumers = group.getConsumers();
        int index = group.getNextConsumerIndex() % consumers.size();
        group.setNextConsumerIndex((group.getNextConsumerIndex() + 1) % consumers.size());
        return index;
    }

    private void markAsDelivered(Message message, String consumerId) {
        message.setStatus(MessageStatus.DELIVERED);
        message.setDeliveredAt(Instant.now());
        message.setDeliveredToConsumerId(consumerId);
    }

    private void registerConsumerHeartbeat(ConsumerGroup group, String consumerId) {
        List<Consumer> consumers = group.getConsumers();
        synchronized (consumers) {
            Consumer consumer = consumers.stream()
                    .filter(c -> c.getId().equals(consumerId))
                    .findFirst()
                    .orElse(null);

            if (consumer == null) {
                consumer = Consumer.builder()
                        .id(consumerId)
                        .groupId(group.getId())
                        .topic(group.getTopic())
                        .lastHeartbeat(Instant.now())
                        .build();
                consumers.add(consumer);
                log.info("Consumer {} joined group {} for topic {}", consumerId, group.getId(), group.getTopic());
            } else {
                consumer.setLastHeartbeat(Instant.now());
            }
        }
    }

    public void moveToDeadLetter(Topic topic, Message message) {
        message.setStatus(MessageStatus.DEAD_LETTER);
        topic.getDeadLetterQueue().add(message);
        log.warn("Message {} moved to dead letter queue for topic {}", message.getId(), topic.getName());
    }

    public boolean resendDeadLetter(String topicName, String messageId) {
        Topic topic = topicManager.getTopic(topicName);
        if (topic == null) {
            return false;
        }

        List<Message> deadLetters = topic.getDeadLetterQueue();
        synchronized (deadLetters) {
            Message target = null;
            int targetIndex = -1;

            for (int i = 0; i < deadLetters.size(); i++) {
                Message msg = deadLetters.get(i);
                if (msg.getId().equals(messageId)) {
                    target = msg;
                    targetIndex = i;
                    break;
                }
            }

            if (target == null) {
                return false;
            }

            deadLetters.remove(targetIndex);

            target.setStatus(MessageStatus.PENDING_DELIVERY);
            target.setRetryCount(0);
            target.setDeliveredAt(null);
            target.setDeliveredToConsumerId(null);
            target.setNextRetryAt(null);
            target.setLastError(null);

            topic.getConsumerGroups().forEach((groupId, group) -> {
                group.setConsumedCount(0);
                group.setLastConsumedIndex(-1);
                group.setNextConsumerIndex(0);
                persistenceService.saveConsumerProgress(topicName, groupId, 0, -1);
            });

            return true;
        }
    }

    private MessageFormat parseFormat(String formatStr) {
        if (formatStr == null || formatStr.isBlank()) {
            return MessageFormat.STRING;
        }
        try {
            return MessageFormat.valueOf(formatStr.toUpperCase());
        } catch (IllegalArgumentException e) {
            throw new IllegalArgumentException("Invalid format: " + formatStr);
        }
    }

    private void validateFormat(String body, MessageFormat format) {
        if (format == MessageFormat.JSON) {
            try {
                objectMapper.readTree(body);
            } catch (Exception e) {
                throw new IllegalArgumentException("Invalid JSON format");
            }
        }
    }
}
