package com.example.lightweightqueue.service;

import com.example.lightweightqueue.model.Consumer;
import com.example.lightweightqueue.model.ConsumerGroup;
import com.example.lightweightqueue.model.DeliveryFailureType;
import com.example.lightweightqueue.model.Message;
import com.example.lightweightqueue.model.MessageStatus;
import com.example.lightweightqueue.model.Topic;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.util.Iterator;
import java.util.List;

@Slf4j
@Component
@RequiredArgsConstructor
public class TimeoutCheckScheduler {

    private static final long ACK_TIMEOUT_SECONDS = 30;
    private static final long CONSUMER_HEARTBEAT_TIMEOUT_SECONDS = 60;

    private final TopicManager topicManager;
    private final MessageQueueService messageQueueService;

    @Scheduled(fixedRate = 5000)
    public void checkTimeouts() {
        for (Topic topic : topicManager.getAllTopics()) {
            for (ConsumerGroup group : topic.getConsumerGroups().values()) {
                checkConsumerTimeouts(group);
                checkMessageTimeouts(topic, group);
            }
        }
    }

    private void checkConsumerTimeouts(ConsumerGroup group) {
        Instant now = Instant.now();
        List<Consumer> consumers = group.getConsumers();
        
        synchronized (consumers) {
            Iterator<Consumer> it = consumers.iterator();
            while (it.hasNext()) {
                Consumer consumer = it.next();
                if (consumer.getLastHeartbeat() != null &&
                        now.isAfter(consumer.getLastHeartbeat().plusSeconds(CONSUMER_HEARTBEAT_TIMEOUT_SECONDS))) {
                    log.info("Consumer {} timeout, removing from group {}", consumer.getId(), group.getId());
                    reassignInFlightMessages(group, consumer.getId());
                    it.remove();
                }
            }
        }
    }

    private void reassignInFlightMessages(ConsumerGroup group, String consumerId) {
        for (Message message : group.getInFlightMessages().values()) {
            if (consumerId.equals(message.getDeliveredToConsumerId())) {
                message.setStatus(MessageStatus.PENDING_DELIVERY);
                message.setDeliveredToConsumerId(null);
                message.setDeliveredAt(null);
                log.info("Reassigning message {} from consumer {}", message.getId(), consumerId);
            }
        }
    }

    private void checkMessageTimeouts(Topic topic, ConsumerGroup group) {
        Instant now = Instant.now();
        Iterator<Message> it = group.getInFlightMessages().values().iterator();
        
        while (it.hasNext()) {
            Message message = it.next();
            if (message.getDeliveredAt() != null &&
                    now.isAfter(message.getDeliveredAt().plusSeconds(ACK_TIMEOUT_SECONDS))) {
                log.warn("Message {} ack timeout, retry count: {}", message.getId(), message.getRetryCount());
                messageQueueService.handleTimeoutOrFailure(topic, group, message, DeliveryFailureType.TIMEOUT);
            }
        }
    }
}
