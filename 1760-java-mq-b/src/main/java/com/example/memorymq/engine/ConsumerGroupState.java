package com.example.memorymq.engine;

import lombok.Builder;
import lombok.Data;

import java.time.Instant;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicLong;

@Data
@Builder
public class ConsumerGroupState {
    private String groupId;
    private AtomicLong lastCommittedOffset;
    private AtomicLong nextDeliveryOffset;
    private AtomicInteger pendingAckCount;
    private Instant lastActivityAt;
}
