package com.servicemesh.common.model;

import java.util.List;
import java.util.Map;

public class TopologyData {
    private long timestamp;
    private List<Node> nodes;
    private List<Edge> edges;

    public long getTimestamp() { return timestamp; }
    public void setTimestamp(long timestamp) { this.timestamp = timestamp; }
    public List<Node> getNodes() { return nodes; }
    public void setNodes(List<Node> nodes) { this.nodes = nodes; }
    public List<Edge> getEdges() { return edges; }
    public void setEdges(List<Edge> edges) { this.edges = edges; }

    public static class Node {
        private String id;
        private String name;
        private String version;
        private String zone;
        private boolean healthy;
        private int instanceCount;
        private long qps;
        private double errorRate;

        public String getId() { return id; }
        public void setId(String id) { this.id = id; }
        public String getName() { return name; }
        public void setName(String name) { this.name = name; }
        public String getVersion() { return version; }
        public void setVersion(String version) { this.version = version; }
        public String getZone() { return zone; }
        public void setZone(String zone) { this.zone = zone; }
        public boolean isHealthy() { return healthy; }
        public void setHealthy(boolean healthy) { this.healthy = healthy; }
        public int getInstanceCount() { return instanceCount; }
        public void setInstanceCount(int instanceCount) { this.instanceCount = instanceCount; }
        public long getQps() { return qps; }
        public void setQps(long qps) { this.qps = qps; }
        public double getErrorRate() { return errorRate; }
        public void setErrorRate(double errorRate) { this.errorRate = errorRate; }
    }

    public static class Edge {
        private String source;
        private String target;
        private long qps;
        private double errorRate;

        public String getSource() { return source; }
        public void setSource(String source) { this.source = source; }
        public String getTarget() { return target; }
        public void setTarget(String target) { this.target = target; }
        public long getQps() { return qps; }
        public void setQps(long qps) { this.qps = qps; }
        public double getErrorRate() { return errorRate; }
        public void setErrorRate(double errorRate) { this.errorRate = errorRate; }
    }
}
