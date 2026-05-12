package com.example.retry.dto;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CreateTaskRequest {
    private String operationId;
    private String payload;
    private Integer maxRetries;
    private Long initialIntervalMs;
    private Long maxIntervalMs;
}
