package com.example.lightweightqueue.service;

import com.example.lightweightqueue.model.Consumer;
import com.example.lightweightqueue.model.ConsumerGroup;
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

    private static final long CONSUMER_HEARTBEAT_TIMEOUT_SECONDS = 60;

    private final TopicManager topicManager;
    private final MessageQueueService messageQueueService;

    @Scheduled(fixedRate = 5000)
    public void checkTimeouts() {
        for (Topic topic : topicManager.getAllTopics()) {
            for (ConsumerGroup group : topic.getConsumerGroups().values()) {
                checkConsumerTimeouts(group);
                messageQueueService.checkAndHandleTimeouts(topic, group);
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
                    messageQueueService.reassignInFlightMessages(group, consumer.getId());
                    it.remove();
                }
            }
        }
    }
}
