package com.example.loadbalancer.dto;

import lombok.Builder;
import lombok.Data;

import java.util.List;

@Data
@Builder
public class NodesListResponse {
    private String currentStrategy;
    private List<NodeResponse> nodes;
}
