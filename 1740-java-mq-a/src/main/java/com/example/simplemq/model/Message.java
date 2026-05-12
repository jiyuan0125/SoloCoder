package com.example.simplemq.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Message {
    private String id;
    private String topic;
    private String content;
    private long timestamp;
    private long offset;

    public Message(String topic, String content) {
        this.id = UUID.randomUUID().toString();
        this.topic = topic;
        this.content = content;
        this.timestamp = System.currentTimeMillis();
    }
}
