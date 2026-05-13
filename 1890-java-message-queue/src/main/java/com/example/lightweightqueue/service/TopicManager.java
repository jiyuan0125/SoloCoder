package com.example.lightweightqueue.service;

import com.example.lightweightqueue.model.Consumer;
import com.example.lightweightqueue.model.ConsumerGroup;
import com.example.lightweightqueue.model.Message;
import com.example.lightweightqueue.model.Topic;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.Collection;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;

@Service
public class TopicManager {

    @Value("${queue.default-max-capacity:10000}")
    private int defaultMaxCapacity = 10000;

    private final ConcurrentMap<String, Topic> topics = new ConcurrentHashMap<>();

    public Topic createTopic(String name) {
        return createTopic(name, defaultMaxCapacity);
    }

    public Topic createTopic(String name, int maxCapacity) {
        return topics.computeIfAbsent(name, k -> new Topic(name, maxCapacity));
    }

    public Topic getTopic(String name) {
        return topics.get(name);
    }

    public Topic getOrCreateTopic(String name) {
        return topics.computeIfAbsent(name, k -> new Topic(name, defaultMaxCapacity));
    }

    public Collection<Topic> getAllTopics() {
        return topics.values();
    }

    public ConsumerGroup getOrCreateConsumerGroup(Topic topic, String groupId) {
        return topic.getConsumerGroups().computeIfAbsent(groupId, k -> new ConsumerGroup(groupId, topic.getName()));
    }

    public ConsumerGroup getConsumerGroup(Topic topic, String groupId) {
        return topic.getConsumerGroups().get(groupId);
    }

    public void removeConsumer(ConsumerGroup group, Consumer consumer) {
        group.getConsumers().removeIf(c -> c.getId().equals(consumer.getId()));
    }
}
