package com.depgraph.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class GraphData {
    private List<NodeInfo> nodes;
    private List<EdgeInfo> edges;

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class NodeInfo {
        private String name;
        private ServiceNode.HealthStatus status;
    }

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class EdgeInfo {
        private String from;
        private String to;
        private long callCount;
    }
}
