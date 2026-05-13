package com.example.lightweightqueue.model;

import lombok.Data;
import lombok.Builder;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Topic {
    private String name;
    private int maxCapacity;
    private List<Message> messages;
    private ConcurrentMap<String, ConsumerGroup> consumerGroups;
    private List<Message> deadLetterQueue;
    private long droppedCount;

    public Topic(String name, int maxCapacity) {
        this.name = name;
        this.maxCapacity = maxCapacity;
        this.messages = new ArrayList<>();
        this.consumerGroups = new ConcurrentHashMap<>();
        this.deadLetterQueue = new ArrayList<>();
        this.droppedCount = 0;
    }

    public synchronized boolean addMessage(Message message) {
        while (messages.size() >= maxCapacity) {
            messages.remove(0);
            droppedCount++;
        }
        messages.add(message);
        return true;
    }
}
