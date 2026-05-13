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
public class ConsumerGroup {
    private String id;
    private String topic;
    private List<Consumer> consumers;
    private long consumedCount;
    private long lastConsumedIndex;
    private ConcurrentMap<String, Message> inFlightMessages;
    private int nextConsumerIndex;

    public ConsumerGroup(String id, String topic) {
        this.id = id;
        this.topic = topic;
        this.consumers = new ArrayList<>();
        this.consumedCount = 0;
        this.lastConsumedIndex = -1;
        this.inFlightMessages = new ConcurrentHashMap<>();
        this.nextConsumerIndex = 0;
    }
}
