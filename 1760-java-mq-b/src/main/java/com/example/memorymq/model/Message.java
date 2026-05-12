package com.example.memorymq.model;

import lombok.Builder;
import lombok.Data;

import java.time.Instant;
import java.util.UUID;

@Data
@Builder
public class Message {
    @Builder.Default
    private String id = UUID.randomUUID().toString();
    private String topic;
    private Object payload;
    @Builder.Default
    private Instant createdAt = Instant.now();
}
