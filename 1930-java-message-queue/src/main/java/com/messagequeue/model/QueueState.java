package com.messagequeue.model;

import lombok.Data;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.concurrent.ConcurrentMap;
import java.util.concurrent.atomic.AtomicLong;

@Data
public class QueueState {
    private final String name;
    private final ConcurrentLinkedQueue<Message> messages = new ConcurrentLinkedQueue<>();
    private final ConcurrentMap<String, AtomicLong> groupOffsets = new ConcurrentHashMap<>();
    private final ConcurrentMap<String, Subscription> subscriptions = new ConcurrentHashMap<>();
    private final List<Message> deadLetters = new ArrayList<>();
    private final Map<String, Message> inFlight = new HashMap<>();

    public QueueState(String name) {
        this.name = name;
    }

    public AtomicLong getOrCreateOffset(String groupId) {
        return groupOffsets.computeIfAbsent(groupId, k -> new AtomicLong(0));
    }

    public void addMessage(Message message) {
        messages.offer(message);
    }

    public void addDeadLetter(Message message) {
        synchronized (deadLetters) {
            deadLetters.add(message);
        }
    }

    public void addSubscription(Subscription subscription) {
        subscriptions.put(subscription.getGroupId(), subscription);
        getOrCreateOffset(subscription.getGroupId());
    }

    public void removeSubscription(String groupId) {
        subscriptions.remove(groupId);
        groupOffsets.remove(groupId);
    }
}
