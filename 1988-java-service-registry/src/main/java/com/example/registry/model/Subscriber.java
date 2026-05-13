package com.example.registry.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Subscriber {
    
    private String subscriberId;
    private String serviceName;
    private String callbackUrl;
    private int consecutiveFailures;
}
