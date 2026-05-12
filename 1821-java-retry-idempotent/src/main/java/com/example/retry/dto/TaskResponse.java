package com.example.retry.dto;

import com.example.retry.model.TaskStatus;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class TaskResponse {
    private String id;
    private String operationId;
    private TaskStatus status;
    private String result;
    private String errorMessage;
    private Integer retryCount;
    private Instant nextRetryTime;
    private Instant createdAt;
    private Instant updatedAt;
}
