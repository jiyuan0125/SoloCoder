package com.messagequeue.service;

import com.messagequeue.filter.FilterExpression;
import com.messagequeue.filter.FilterParser;
import com.messagequeue.model.Message;
import com.messagequeue.model.QueueState;
import com.messagequeue.model.Subscription;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.ArrayList;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.atomic.AtomicLong;

@Slf4j
@Service
public class QueueService {

    @Value("${queue.max-retry:3}")
    private int maxRetry;

    @Value("${queue.dead-letter-queue-suffix:-dlq}")
    private String deadLetterQueueSuffix;

    private final ConcurrentMap<String, QueueState> queues = new ConcurrentHashMap<>();

    public Subscription subscribe(String queueName, String groupId, String filterExpression) {
        if (filterExpression != null && !filterExpression.trim().isEmpty()) {
            try {
                FilterParser.parse(filterExpression);
            } catch (Exception e) {
                throw new IllegalArgumentException("Invalid filter expression: " + e.getMessage(), e);
            }
        }

        QueueState queue = queues.computeIfAbsent(queueName, k -> new QueueState(k));
        Instant now = Instant.now();

        Subscription existing = queue.getSubscriptions().get(groupId);
        Subscription subscription;
        if (existing != null) {
            subscription = Subscription.builder()
                    .queueName(queueName)
                    .groupId(groupId)
                    .filterExpression(filterExpression)
                    .createdAt(existing.getCreatedAt())
                    .updatedAt(now)
                    .build();
            log.info("Updated subscription: queue={}, group={}", queueName, groupId);
        } else {
            subscription = Subscription.builder()
                    .queueName(queueName)
                    .groupId(groupId)
                    .filterExpression(filterExpression)
                    .createdAt(now)
                    .updatedAt(now)
                    .build();
            log.info("Created subscription: queue={}, group={}", queueName, groupId);
        }

        queue.addSubscription(subscription);
        return subscription;
    }

    public boolean unsubscribe(String queueName, String groupId) {
        QueueState queue = queues.get(queueName);
        if (queue == null) {
            return false;
        }
        boolean removed = queue.getSubscriptions().containsKey(groupId);
        queue.removeSubscription(groupId);
        if (removed) {
            log.info("Removed subscription: queue={}, group={}", queueName, groupId);
        }
        return removed;
    }

    public List<Subscription> listSubscriptions(String queueName) {
        QueueState queue = queues.get(queueName);
        if (queue == null) {
            return List.of();
        }
        return new ArrayList<>(queue.getSubscriptions().values());
    }

    public void publish(String queueName, Map<String, String> headers, String body) {
        QueueState queue = queues.computeIfAbsent(queueName, k -> new QueueState(k));
        Message message = Message.create(headers, body);
        queue.addMessage(message);
        log.debug("Published message: queue={}, id={}", queueName, message.getId());
    }

    public Message consume(String queueName, String groupId) {
        QueueState queue = queues.get(queueName);
        if (queue == null) {
            return null;
        }

        Subscription subscription = queue.getSubscriptions().get(groupId);
        if (subscription == null) {
            throw new IllegalArgumentException("No subscription found for group: " + groupId);
        }

        FilterExpression filter;
        try {
            filter = FilterParser.parse(subscription.getFilterExpression());
        } catch (Exception e) {
            log.error("Failed to parse filter expression, using pass-through", e);
            filter = m -> true;
        }

        AtomicLong offset = queue.getOrCreateOffset(groupId);
        ConcurrentLinkedQueue<Message> messages = queue.getMessages();
        Iterator<Message> iterator = messages.iterator();
        long currentOffset = offset.get();

        int skipped = 0;
        while (iterator.hasNext()) {
            Message message = iterator.next();
            if (skipped < currentOffset) {
                skipped++;
                continue;
            }

            if (filter.evaluate(message)) {
                synchronized (queue.getInFlight()) {
                    queue.getInFlight().put(message.getId(), message);
                }
                offset.incrementAndGet();
                log.debug("Consumed message: queue={}, group={}, id={}", queueName, groupId, message.getId());
                return message;
            } else {
                offset.incrementAndGet();
                log.debug("Filtered message: queue={}, group={}, id={}", queueName, groupId, message.getId());
            }
        }

        return null;
    }

    public boolean acknowledge(String queueName, String groupId, String messageId) {
        QueueState queue = queues.get(queueName);
        if (queue == null) {
            return false;
        }

        Message message;
        synchronized (queue.getInFlight()) {
            message = queue.getInFlight().remove(messageId);
        }

        if (message == null) {
            return false;
        }

        removeFromQueue(queue, message);
        log.debug("Acknowledged message: queue={}, group={}, id={}", queueName, groupId, messageId);
        return true;
    }

    public boolean nack(String queueName, String groupId, String messageId) {
        QueueState queue = queues.get(queueName);
        if (queue == null) {
            return false;
        }

        Message message;
        synchronized (queue.getInFlight()) {
            message = queue.getInFlight().remove(messageId);
        }

        if (message == null) {
            return false;
        }

        message.incrementRetry();
        if (message.getRetryCount() >= maxRetry) {
            moveToDeadLetter(queue, message);
            log.warn("Message moved to dead letter: queue={}, id={}, retries={}", queueName, message.getId(), message.getRetryCount());
        } else {
            queue.addMessage(message);
            log.debug("Message requeued: queue={}, id={}, retry={}", queueName, message.getId(), message.getRetryCount());
        }

        return true;
    }

    private void removeFromQueue(QueueState queue, Message message) {
        queue.getMessages().removeIf(m -> m.getId().equals(message.getId()));
    }

    private void moveToDeadLetter(QueueState queue, Message message) {
        queue.addDeadLetter(message);
        removeFromQueue(queue, message);
    }

    public List<Message> getDeadLetters(String queueName) {
        QueueState queue = queues.get(queueName);
        if (queue == null) {
            return List.of();
        }
        synchronized (queue.getDeadLetters()) {
            return new ArrayList<>(queue.getDeadLetters());
        }
    }

    public Message consumeDeadLetter(String queueName, String groupId) {
        QueueState queue = queues.get(queueName);
        if (queue == null) {
            return null;
        }

        Subscription subscription = queue.getSubscriptions().get(groupId);
        if (subscription == null) {
            throw new IllegalArgumentException("No subscription found for group: " + groupId);
        }

        FilterExpression filter;
        try {
            filter = FilterParser.parse(subscription.getFilterExpression());
        } catch (Exception e) {
            filter = m -> true;
        }

        synchronized (queue.getDeadLetters()) {
            Iterator<Message> iterator = queue.getDeadLetters().iterator();
            while (iterator.hasNext()) {
                Message message = iterator.next();
                if (filter.evaluate(message)) {
                    iterator.remove();
                    log.debug("Consumed dead letter: queue={}, group={}, id={}", queueName, groupId, message.getId());
                    return message;
                }
            }
        }

        return null;
    }

    public QueueState getQueueState(String queueName) {
        return queues.get(queueName);
    }
}
