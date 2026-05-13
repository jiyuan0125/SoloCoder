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
public class Consumer {
    private String id;
    private String groupId;
    private String topic;
    private Instant lastHeartbeat;
}
