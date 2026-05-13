package com.example.lightweightqueue.service;

import com.example.lightweightqueue.model.Consumer;
import com.example.lightweightqueue.model.ConsumerGroup;
import com.example.lightweightqueue.model.DeliveryFailureType;
import com.example.lightweightqueue.model.GroupMessageState;
import com.example.lightweightqueue.model.Message;
import com.example.lightweightqueue.model.MessageFormat;
import com.example.lightweightqueue.model.MessageStatus;
import com.example.lightweightqueue.model.Topic;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.Iterator;
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

        GroupMessageState state = group.getMessageState(messageId);
        if (state != null) {
            state.setStatus(MessageStatus.ACKNOWLEDGED);
        }

        group.setConsumedCount(group.getConsumedCount() + 1);

        persistenceService.saveConsumerProgress(topicName, groupId, group.getConsumedCount(), group.getLastConsumedIndex());

        return true;
    }

    public void handleTimeoutOrFailure(Topic topic, ConsumerGroup group, String messageId, DeliveryFailureType failureType) {
        GroupMessageState state = group.getMessageState(messageId);
        if (state == null) {
            return;
        }

        state.setLastError(failureType.name());

        if (state.getRetryCount() >= MAX_RETRY_COUNT) {
            moveToDeadLetter(topic, group, messageId);
            return;
        }

        long retryInterval = RETRY_INTERVALS[Math.min(state.getRetryCount(), RETRY_INTERVALS.length - 1)];
        state.setRetryCount(state.getRetryCount() + 1);
        state.setStatus(MessageStatus.PENDING_DELIVERY);
        state.setNextRetryAt(Instant.now().plusSeconds(retryInterval));
        state.setDeliveredToConsumerId(null);
        state.setDeliveredAt(null);

        group.getInFlightMessages().remove(messageId);
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

                GroupMessageState state = group.getOrCreateMessageState(msg.getId());

                if (state.getStatus() == MessageStatus.DEAD_LETTER || state.getStatus() == MessageStatus.ACKNOWLEDGED) {
                    continue;
                }

                if (state.getStatus() == MessageStatus.DELIVERED) {
                    continue;
                }

                if (state.getNextRetryAt() != null && state.getNextRetryAt().isAfter(now)) {
                    continue;
                }

                if (group.getInFlightMessages().containsKey(msg.getId())) {
                    continue;
                }

                int assignedIndex = assignConsumerIndex(group, msg);
                Consumer assignedConsumer = consumers.get(assignedIndex);

                if (assignedConsumer.getId().equals(consumerId)) {
                    markAsDelivered(group, msg, consumerId);
                    group.setLastConsumedIndex(i);
                    group.getInFlightMessages().put(msg.getId(), msg);
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

    private void markAsDelivered(ConsumerGroup group, Message message, String consumerId) {
        GroupMessageState state = group.getOrCreateMessageState(message.getId());
        state.setStatus(MessageStatus.DELIVERED);
        state.setDeliveredAt(Instant.now());
        state.setDeliveredToConsumerId(consumerId);
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

    public void moveToDeadLetter(Topic topic, ConsumerGroup group, String messageId) {
        GroupMessageState state = group.getMessageState(messageId);
        if (state != null) {
            state.setStatus(MessageStatus.DEAD_LETTER);
        }

        Message originalMessage = topic.getMessages().stream()
                .filter(m -> m.getId().equals(messageId))
                .findFirst()
                .orElse(null);

        if (originalMessage != null) {
            Message deadLetterCopy = Message.builder()
                    .id(originalMessage.getId() + "-" + group.getId() + "-" + UUID.randomUUID())
                    .topic(originalMessage.getTopic())
                    .body(originalMessage.getBody())
                    .format(originalMessage.getFormat())
                    .status(MessageStatus.DEAD_LETTER)
                    .createdAt(Instant.now())
                    .lastError("DEAD_LETTER_FROM_GROUP_" + group.getId())
                    .build();
            
            topic.getDeadLetterQueue().add(deadLetterCopy);
            log.warn("Message {} (group {}) moved to dead letter queue for topic {}", 
                    messageId, group.getId(), topic.getName());
        }

        group.getInFlightMessages().remove(messageId);
    }

    public boolean resendDeadLetter(String topicName, String deadLetterMessageId) {
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
                if (msg.getId().equals(deadLetterMessageId)) {
                    target = msg;
                    targetIndex = i;
                    break;
                }
            }

            if (target == null) {
                return false;
            }

            deadLetters.remove(targetIndex);

            String newMessageId = UUID.randomUUID().toString();
            Message newMessage = Message.builder()
                    .id(newMessageId)
                    .topic(topicName)
                    .body(target.getBody())
                    .format(target.getFormat())
                    .status(MessageStatus.PENDING_DELIVERY)
                    .createdAt(Instant.now())
                    .build();

            topic.addMessage(newMessage);

            log.info("Dead letter message {} resent as new message {} for topic {}", 
                    deadLetterMessageId, newMessageId, topicName);

            return true;
        }
    }

    public void checkAndHandleTimeouts(Topic topic, ConsumerGroup group) {
        Instant now = Instant.now();
        List<String> toHandle = new java.util.ArrayList<>();

        for (GroupMessageState state : group.getMessageStates().values()) {
            if (state.getStatus() == MessageStatus.DELIVERED &&
                    state.getDeliveredAt() != null &&
                    now.isAfter(state.getDeliveredAt().plusSeconds(ACK_TIMEOUT_SECONDS))) {
                log.warn("Message {} ack timeout in group {}, retry count: {}", 
                        state.getMessageId(), group.getId(), state.getRetryCount());
                toHandle.add(state.getMessageId());
            }
        }

        for (String messageId : toHandle) {
            handleTimeoutOrFailure(topic, group, messageId, DeliveryFailureType.TIMEOUT);
        }
    }

    public void reassignInFlightMessages(ConsumerGroup group, String consumerId) {
        Iterator<String> it = group.getInFlightMessages().keySet().iterator();
        while (it.hasNext()) {
            String messageId = it.next();
            GroupMessageState state = group.getMessageState(messageId);
            if (state != null && consumerId.equals(state.getDeliveredToConsumerId())) {
                state.setStatus(MessageStatus.PENDING_DELIVERY);
                state.setDeliveredToConsumerId(null);
                state.setDeliveredAt(null);
                log.info("Reassigning message {} from consumer {} in group {}", 
                        messageId, consumerId, group.getId());
            }
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
