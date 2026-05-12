package com.example.circuitbreaker.api;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class BreakerStatus {
    private String serviceName;
    private String state;
    private Integer failureThreshold;
    private Integer openDurationSeconds;
    private List<CallRecordView> recentCalls;
}
