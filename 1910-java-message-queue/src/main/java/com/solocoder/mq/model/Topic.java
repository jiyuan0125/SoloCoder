package com.solocoder.mq.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;

@Data
@NoArgsConstructor
public class Topic {

    @JsonProperty("name")
    private String name;

    @JsonProperty("capacity")
    private int capacity;

    @JsonProperty("next_offset")
    private long nextOffset;

    @JsonProperty("created_at")
    private long createdAt;

    private transient List<Message> messages = new ArrayList<>();

    private transient ConcurrentMap<String, ConsumerGroup> consumerGroups = new ConcurrentHashMap<>();

    public static Topic create(String name, int capacity) {
        Topic topic = new Topic();
        topic.setName(name);
        topic.setCapacity(capacity);
        topic.setNextOffset(0);
        topic.setCreatedAt(System.currentTimeMillis());
        return topic;
    }
}
