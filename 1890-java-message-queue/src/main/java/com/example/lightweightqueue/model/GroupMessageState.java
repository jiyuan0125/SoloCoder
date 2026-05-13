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
public class GroupMessageState {
    private String messageId;
    private MessageStatus status;
    private int retryCount;
    private Instant deliveredAt;
    private Instant nextRetryAt;
    private String deliveredToConsumerId;
    private String lastError;
}
