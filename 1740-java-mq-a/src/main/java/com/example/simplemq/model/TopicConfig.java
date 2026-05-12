package com.example.simplemq.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class TopicConfig {
    private String name;
    private long maxMessages;
    private long retentionMs;
    private long createdAt;
    private long updatedAt;

    public TopicConfig(String name) {
        this.name = name;
        this.maxMessages = 1000;
        this.retentionMs = 86400000;
        this.createdAt = System.currentTimeMillis();
        this.updatedAt = System.currentTimeMillis();
    }

    public TopicConfig(String name, long maxMessages, long retentionMs) {
        this.name = name;
        this.maxMessages = maxMessages;
        this.retentionMs = retentionMs;
        this.createdAt = System.currentTimeMillis();
        this.updatedAt = System.currentTimeMillis();
    }
}
