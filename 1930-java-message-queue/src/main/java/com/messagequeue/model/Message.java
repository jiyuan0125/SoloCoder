package com.messagequeue.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Message {
    private String id;
    private Map<String, String> headers;
    private String body;
    private Instant timestamp;
    private int retryCount;

    public static Message create(Map<String, String> headers, String body) {
        return Message.builder()
                .id(UUID.randomUUID().toString())
                .headers(headers != null ? new HashMap<>(headers) : new HashMap<>())
                .body(body)
                .timestamp(Instant.now())
                .retryCount(0)
                .build();
    }

    public Message incrementRetry() {
        this.retryCount++;
        return this;
    }
}
