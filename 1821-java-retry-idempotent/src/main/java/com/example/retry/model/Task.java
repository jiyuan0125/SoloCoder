package com.example.retry.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Task {
    private String id;
    private String operationId;
    private TaskStatus status;
    private String result;
    private String errorMessage;
    private int retryCount;
    private int maxRetries;
    private long initialIntervalMs;
    private long maxIntervalMs;
    private Instant nextRetryTime;
    private Instant createdAt;
    private Instant updatedAt;
}
