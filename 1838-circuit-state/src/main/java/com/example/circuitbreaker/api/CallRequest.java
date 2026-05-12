package com.example.circuitbreaker.api;

import lombok.Data;

@Data
public class CallRequest {
    private String requestId;
    private boolean success = true;
}
