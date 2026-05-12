package com.example.circuitbreaker.stats;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CallRecord {
    private String serviceName;
    private String requestId;
    private boolean success;
    private Instant timestamp;
}
