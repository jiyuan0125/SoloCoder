package com.example.loadbalancer.dto;

import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class NodeResponse {
    private String id;
    private String ip;
    private int port;
    private int initialWeight;
    private int currentWeight;
    private String status;
    private int activeConnections;
    private long registeredAt;
}
