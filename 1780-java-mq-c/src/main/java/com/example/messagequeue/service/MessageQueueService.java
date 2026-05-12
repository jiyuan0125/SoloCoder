package com.example.messagequeue.service;

import com.example.messagequeue.config.MessageQueueProperties;
import com.example.messagequeue.exception.QueueFullException;
import com.example.messagequeue.exception.TopicNotFoundException;
import com.example.messagequeue.model.Message;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import reactor.core.publisher.Mono;
import reactor.core.scheduler.Schedulers;

import java.time.Duration;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.atomic.AtomicLong;
import java.util.concurrent.locks.Condition;
import java.util.concurrent.locks.ReentrantLock;

@Slf4j
@Service
@RequiredArgsConstructor
public class MessageQueueService {

    private final MessageQueueProperties properties;
    private final OffsetService offsetService;

    private final Map<String, TopicQueue> topicQueues = new ConcurrentHashMap<>();
    private final Map<String, AtomicLong> topicOffsets = new ConcurrentHashMap<>();

    public Mono<Message> send(String topic, String content, String contentType) {
        TopicQueue queue = getOrCreateTopicQueue(topic);
        long offset = getNextOffset(topic);
        Message message = new Message(
                java.util.UUID.randomUUID().toString(),
                topic,
                content,
                contentType,
                System.currentTimeMillis(),
                offset
        );
        return tryEnqueue(queue, message);
    }

    private Mono<Message> tryEnqueue(TopicQueue queue, Message message) {
        return Mono.create(sink -> {
            int maxSize = properties.getMaxQueueSize();
            int timeoutSeconds = properties.getWaitTimeoutSeconds();
            long deadline = System.currentTimeMillis() + timeoutSeconds * 1000L;
            Schedulers.boundedElastic().schedule(() -> {
                try {
                    boolean success = false;
                    while (!success) {
                        queue.lock.lock();
                        try {
                            if (queue.queue.size() < maxSize) {
                                queue.queue.offer(message);
                                queue.notFull.signalAll();
                                success = true;
                                sink.success(message);
                                log.debug("Enqueued message to topic {}: offset={}", message.getTopic(), message.getOffset());
                            } else {
                                long remaining = deadline - System.currentTimeMillis();
                                if (remaining <= 0) {
                                    throw new QueueFullException(
                                            message.getTopic(),
                                            queue.queue.size(),
                                            maxSize,
                                            timeoutSeconds
                                    );
                                }
                                log.debug("Queue full for topic {}, waiting... (remaining={}ms)", message.getTopic(), remaining);
                                queue.notFull.await(remaining, java.util.concurrent.TimeUnit.MILLISECONDS);
                            }
                        } finally {
                            queue.lock.unlock();
                        }
                    }
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    sink.error(e);
                } catch (QueueFullException e) {
                    sink.error(e);
                }
            });
        });
    }

    public Message receive(String topic) {
        TopicQueue queue = topicQueues.get(topic);
        if (queue == null || queue.queue.isEmpty()) {
            return null;
        }
        queue.lock.lock();
        try {
            Message message = queue.queue.poll();
            if (message != null) {
                queue.notFull.signalAll();
                log.debug("Dequeued message from topic {}: offset={}", topic, message.getOffset());
            }
            return message;
        } finally {
            queue.lock.unlock();
        }
    }

    public void commitOffset(String topic, long offset, String consumerId) {
        if (!topicQueues.containsKey(topic)) {
            throw new TopicNotFoundException(topic);
        }
        offsetService.persistOffset(topic, consumerId, offset);
        log.info("Committed offset for topic {}, consumer {}: {}", topic, consumerId, offset);
    }

    public long getLastCommittedOffset(String topic, String consumerId) {
        return offsetService.getLastCommittedOffset(topic, consumerId);
    }

    private TopicQueue getOrCreateTopicQueue(String topic) {
        return topicQueues.computeIfAbsent(topic, t -> new TopicQueue());
    }

    private long getNextOffset(String topic) {
        return topicOffsets.computeIfAbsent(topic, t -> new AtomicLong(0)).incrementAndGet();
    }

    private static class TopicQueue {
        final LinkedBlockingQueue<Message> queue = new LinkedBlockingQueue<>();
        final ReentrantLock lock = new ReentrantLock();
        final Condition notFull = lock.newCondition();
    }
}
