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
public class StateChangeEvent {
    private String serviceName;
    private String oldState;
    private String newState;
    private ChangeReason reason;
    private String requestId;
    private Instant timestamp;
}
