package com.example.memorymq.engine;

import com.example.memorymq.config.MqProperties;
import com.example.memorymq.model.Message;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.util.*;
import java.util.concurrent.*;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicLong;

@Slf4j
public class TopicPartition {
    private final String topic;
    private final int maxMessages;
    private final MqProperties properties;
    private final ConcurrentSkipListMap<Long, Message> messageStore;
    private final AtomicLong nextOffset;
    private final AtomicLong baseOffset;
    private final ConcurrentHashMap<String, ConsumerGroupState> consumerGroups;
    private final ConcurrentHashMap<String, InFlightMessage> inFlightByMessageId;
    private final DelayQueue<InFlightMessage> retryQueue;
    private final AtomicBoolean running;
    private final ScheduledExecutorService scheduler;
    private final Object readLock;

    public TopicPartition(String topic, int maxMessages, MqProperties properties) {
        this.topic = topic;
        this.maxMessages = maxMessages;
        this.properties = properties;
        this.messageStore = new ConcurrentSkipListMap<>();
        this.nextOffset = new AtomicLong(0);
        this.baseOffset = new AtomicLong(0);
        this.consumerGroups = new ConcurrentHashMap<>();
        this.inFlightByMessageId = new ConcurrentHashMap<>();
        this.retryQueue = new DelayQueue<>();
        this.running = new AtomicBoolean(true);
        this.scheduler = Executors.newSingleThreadScheduledExecutor(r -> {
            Thread t = new Thread(r, "topic-" + topic + "-retry-thread");
            t.setDaemon(true);
            return t;
        });
        this.readLock = new Object();
        startRetryTask();
    }

    private void startRetryTask() {
        scheduler.scheduleWithFixedDelay(() -> {
            try {
                processRetryQueue();
                checkTimeoutMessages();
            } catch (Exception e) {
                log.error("Retry task error for topic {}: {}", topic, e.getMessage());
            }
        }, properties.getRetryIntervalSeconds(), properties.getRetryIntervalSeconds(), TimeUnit.SECONDS);
    }

    private void processRetryQueue() {
        InFlightMessage inflight;
        while ((inflight = retryQueue.poll()) != null) {
            if (!running.get()) break;
            String messageId = inflight.getMessageId();
            InFlightMessage current = inFlightByMessageId.remove(messageId);
            if (current != null && current.equals(inflight)) {
                log.debug("Retry message {} available for topic {}", messageId, topic);
            }
        }
    }

    private void checkTimeoutMessages() {
        long now = System.nanoTime();
        long timeoutNanos = TimeUnit.SECONDS.toNanos(properties.getAckTimeoutSeconds());
        List<InFlightMessage> timedOut = new ArrayList<>();
        inFlightByMessageId.entrySet().removeIf(entry -> {
            InFlightMessage ifm = entry.getValue();
            long elapsed = now - TimeUnit.MILLISECONDS.toNanos(
                java.time.Duration.between(ifm.getCreatedAt(), Instant.now()).toMillis());
            long elapsedNanos = java.time.Duration.between(ifm.getCreatedAt(), Instant.now()).toNanos();
            if (elapsedNanos > timeoutNanos) {
                timedOut.add(ifm);
                return true;
            }
            return false;
        });
        for (InFlightMessage ifm : timedOut) {
            handleTimeout(ifm);
        }
    }

    private void handleTimeout(InFlightMessage ifm) {
        if (ifm.getRetryCount() >= properties.getDefaultMaxRetry()) {
            log.warn("Message {} exceeded max retries, dropping", ifm.getMessageId());
            return;
        }
        InFlightMessage requeued = new InFlightMessage(
            ifm.getMessageId(), ifm.getMessage(), ifm.getConsumerId(),
            ifm.getRetryCount() + 1,
            TimeUnit.SECONDS.toMillis(properties.getRetryIntervalSeconds() * (ifm.getRetryCount() + 1)));
        retryQueue.offer(requeued);
    }

    public boolean publish(Message message) {
        synchronized (readLock) {
            long currentSize = messageStore.size();
            if (currentSize >= maxMessages) {
                trimOldMessages();
                if (messageStore.size() >= maxMessages) {
                    return false;
                }
            }
            long offset = nextOffset.getAndIncrement();
            messageStore.put(offset, message);
            return true;
        }
    }

    private void trimOldMessages() {
        Instant cutoff = Instant.now().minus(java.time.Duration.ofMinutes(properties.getRetentionMinutes()));
        while (!messageStore.isEmpty() == false) {
            Map.Entry<Long, Message> entry = messageStore.firstEntry();
            if (entry == null) break;
            Message oldest = entry.getValue();
            if (oldest.getCreatedAt().isBefore(cutoff) || messageStore.size() > maxMessages * 0.8) {
                messageStore.pollFirstEntry();
                baseOffset.set(entry.getKey() + 1);
            } else {
                break;
            }
        }
    }

    public List<Message> consume(String groupId, int maxMessages, long timeoutMillis) {
        ConsumerGroupState state = getOrCreateConsumerGroup(groupId);
        List<Message> result = new ArrayList<>();
        long deadline = System.currentTimeMillis() + timeoutMillis;
        int fetched = 0;
        while (fetched < maxMessages && System.currentTimeMillis() < deadline) {
            synchronized (readLock) {
                Long next = state.getNextDeliveryOffset().get();
                Map.Entry<Long, Message> entry = messageStore.ceilingEntry(next);
                if (entry != null) {
                    long offset = entry.getKey();
                    Message msg = entry.getValue();
                    InFlightMessage inflight = new InFlightMessage(
                        msg.getId(), msg, groupId, 0,
                        TimeUnit.SECONDS.toMillis(properties.getAckTimeoutSeconds()));
                    inFlightByMessageId.put(msg.getId(), inflight);
                    state.getNextDeliveryOffset().set(offset + 1);
                    state.getPendingAckCount().incrementAndGet();
                    state.setLastActivityAt(Instant.now());
                    result.add(msg);
                    fetched++;
                } else {
                    break;
                }
            }
        }
        return result;
    }

    public boolean ack(String groupId, List<String> messageIds) {
        ConsumerGroupState state = consumerGroups.get(groupId);
        if (state == null) return false;
        for (String msgId : messageIds) {
            inFlightByMessageId.remove(msgId);
            state.getPendingAckCount().decrementAndGet();
        }
        state.setLastActivityAt(Instant.now());
        return true;
    }

    public boolean nack(String groupId, List<String> messageIds, boolean requeue) {
        ConsumerGroupState state = consumerGroups.get(groupId);
        if (state == null) return false;
        for (String msgId : messageIds) {
            InFlightMessage ifm = inFlightByMessageId.remove(msgId);
            if (ifm == null) continue;
            state.getPendingAckCount().decrementAndGet();
            if (requeue && ifm.getRetryCount() < properties.getDefaultMaxRetry()) {
                InFlightMessage requeued = new InFlightMessage(
                    ifm.getMessageId(), ifm.getMessage(), groupId,
                    ifm.getRetryCount() + 1,
                    TimeUnit.SECONDS.toMillis(properties.getRetryIntervalSeconds()));
                retryQueue.offer(requeued);
            }
        }
        state.setLastActivityAt(Instant.now());
        return true;
    }

    private ConsumerGroupState getOrCreateConsumerGroup(String groupId) {
        return consumerGroups.computeIfAbsent(groupId, id ->
            ConsumerGroupState.builder()
                .groupId(id)
                .lastCommittedOffset(new AtomicLong(0))
                .nextDeliveryOffset(new AtomicLong(messageStore.isEmpty() ? 0 : messageStore.firstKey()))
                .pendingAckCount(new AtomicInteger(0))
                .lastActivityAt(Instant.now())
                .build()
        );
    }

    public void cleanupExpired() {
        trimOldMessages();
    }

    public int getCurrentMessageCount() {
        return messageStore.size();
    }

    public Set<String> getConsumerGroups() {
        return new HashSet<>(consumerGroups.keySet());
    }

    public void shutdown() {
        running.set(false);
        scheduler.shutdownNow();
    }
}
