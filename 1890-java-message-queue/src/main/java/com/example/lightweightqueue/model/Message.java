package com.example.lightweightqueue.model;

import lombok.Data;
import lombok.Builder;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

import java.time.Instant;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Message {
    private String id;
    private String topic;
    private String body;
    private MessageFormat format;
    private int retryCount;
    private MessageStatus status;
    private Instant createdAt;
    private Instant deliveredAt;
    private Instant nextRetryAt;
    private String deliveredToConsumerId;
    private String lastError;
}
