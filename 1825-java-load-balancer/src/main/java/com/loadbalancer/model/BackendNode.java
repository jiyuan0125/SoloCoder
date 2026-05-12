package com.loadbalancer.model;

import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class BackendNode {
    private String id;
    private String host;
    private int port;
    private int weight;
    @Builder.Default
    private List<String> tags = new ArrayList<>();
    private NodeStatus status;
    private LocalDateTime registeredAt;
    private LocalDateTime statusChangedAt;

    public enum NodeStatus {
        HEALTHY,
        UNHEALTHY,
        DRAINING
    }
}
