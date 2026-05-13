package com.example.proxy.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AccessLog {
    
    private String id;
    private String traceId;
    private LocalDateTime requestTime;
    private String sourceIp;
    private String requestPath;
    private String method;
    private String targetBackend;
    private int responseStatus;
    private long durationMs;
    private String errorMessage;
}
