package com.messagequeue.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Subscription {
    private String queueName;
    private String groupId;
    private String filterExpression;
    private Instant createdAt;
    private Instant updatedAt;
}
