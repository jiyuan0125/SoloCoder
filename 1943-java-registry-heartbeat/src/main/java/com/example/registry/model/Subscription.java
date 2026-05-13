package com.example.registry.model;

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
    private String id;
    private String serviceName;
    private String callbackUrl;
    private Instant subscribeTime;
}
