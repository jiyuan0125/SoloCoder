package com.example.registry.dto;

import com.example.registry.model.ServiceInstance;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.Instant;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class InstanceResponse {
    private String id;
    private String ip;
    private int port;
    private int weight;
    private String healthStatus;
    private Instant lastHeartbeat;
    private Instant registerTime;

    public static InstanceResponse from(ServiceInstance instance) {
        return InstanceResponse.builder()
                .id(instance.getId())
                .ip(instance.getIp())
                .port(instance.getPort())
                .weight(instance.getWeight())
                .healthStatus(instance.getHealthStatus() == ServiceInstance.HealthStatus.HEALTHY ? "healthy" : "unhealthy")
                .lastHeartbeat(instance.getLastHeartbeat())
                .registerTime(instance.getRegisterTime())
                .build();
    }
}
