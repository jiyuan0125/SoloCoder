package com.example.memorymq.model;

import lombok.Builder;
import lombok.Data;

import java.time.Instant;
import java.util.Set;

@Data
@Builder
public class TopicInfo {
    private String name;
    private int maxMessages;
    private int currentMessageCount;
    private Set<String> consumerGroups;
    private Instant createdAt;
}
