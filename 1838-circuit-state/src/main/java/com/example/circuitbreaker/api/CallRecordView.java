package com.example.circuitbreaker.api;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class CallRecordView {
    private String requestId;
    private boolean success;
    private Instant timestamp;
}
