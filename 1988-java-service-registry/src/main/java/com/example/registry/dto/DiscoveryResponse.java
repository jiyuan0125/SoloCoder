package com.example.registry.dto;

import com.example.registry.model.ServiceInstance;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class DiscoveryResponse {
    
    private List<ServiceInstance> instances;
    private int total;
    private int page;
    private int pageSize;
    private int totalPages;
}
