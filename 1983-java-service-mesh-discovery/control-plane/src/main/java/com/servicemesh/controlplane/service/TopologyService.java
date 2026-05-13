package com.servicemesh.controlplane.service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.servicemesh.common.model.ServiceInstance;
import com.servicemesh.common.model.TopologyData;
import com.servicemesh.controlplane.config.ControlPlaneProperties;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;

@Service
public class TopologyService {

    private static final Logger log = LoggerFactory.getLogger(TopologyService.class);

    private final ControlPlaneProperties properties;
    private final OkHttpClient httpClient;
    private final ObjectMapper objectMapper;

    private final Map<String, ServiceNode> nodes = new ConcurrentHashMap<>();
    private final Map<String, ServiceEdgeData> edges = new ConcurrentHashMap<>();
    private final Map<String, List<ServiceInstance>> serviceInstances = new ConcurrentHashMap<>();

    private volatile TopologyData cachedTopology = new TopologyData();

    public TopologyService(ControlPlaneProperties properties) {
        this.properties = properties;
        this.httpClient = new OkHttpClient.Builder()
                .connectTimeout(5, TimeUnit.SECONDS)
                .readTimeout(5, TimeUnit.SECONDS)
                .build();
        this.objectMapper = new ObjectMapper();
    }

    @PostConstruct
    public void init() {
        refreshInstances();
        buildTopology();
    }

    public void recordTelemetry(String source, String target, double qps,
                                 long totalRequests, long errorCount,
                                 double errorRate, double avgLatencyMs) {
        String edgeKey = source + "->" + target;
        ServiceEdgeData edge = edges.computeIfAbsent(edgeKey, k -> new ServiceEdgeData(source, target));
        edge.update(qps, totalRequests, errorCount, errorRate);
    }

    @Scheduled(fixedRateString = "${control-plane.topology-refresh-interval:10000}")
    public void refreshAndBuild() {
        refreshInstances();
        buildTopology();
    }

    private void refreshInstances() {
        try {
            Request request = new Request.Builder()
                    .url(properties.getRegistryUrl() + "/api/registry/instances?healthy=false")
                    .get()
                    .build();

            try (Response response = httpClient.newCall(request).execute()) {
                if (response.isSuccessful() && response.body() != null) {
                    JsonNode root = objectMapper.readTree(response.body().string());
                    serviceInstances.clear();
                    nodes.clear();

                    for (JsonNode node : root) {
                        ServiceInstance inst = objectMapper.treeToValue(node, ServiceInstance.class);
                        String serviceName = inst.getServiceName();
                        serviceInstances.computeIfAbsent(serviceName, k -> new ArrayList<>()).add(inst);

                        ServiceNode svcNode = nodes.computeIfAbsent(serviceName, k -> new ServiceNode(serviceName));
                        svcNode.addInstance(inst);
                    }
                }
            }
        } catch (Exception e) {
            log.warn("Failed to refresh instances: {}", e.getMessage());
        }
    }

    private void buildTopology() {
        TopologyData data = new TopologyData();
        data.setTimestamp(System.currentTimeMillis());

        List<TopologyData.Node> nodeList = new ArrayList<>();
        for (Map.Entry<String, ServiceNode> entry : nodes.entrySet()) {
            ServiceNode sn = entry.getValue();
            TopologyData.Node n = new TopologyData.Node();
            n.setId(sn.getName());
            n.setName(sn.getName());
            n.setHealthy(sn.isHealthy());
            n.setInstanceCount(sn.getInstanceCount());
            n.setQps(sn.getTotalQps());
            n.setErrorRate(sn.getErrorRate());
            nodeList.add(n);
        }
        data.setNodes(nodeList);

        List<TopologyData.Edge> edgeList = new ArrayList<>();
        for (ServiceEdgeData edge : edges.values()) {
            TopologyData.Edge e = new TopologyData.Edge();
            e.setSource(edge.getSource());
            e.setTarget(edge.getTarget());
            e.setQps((long) edge.getQps());
            e.setErrorRate(edge.getErrorRate());
            edgeList.add(e);
        }
        data.setEdges(edgeList);

        this.cachedTopology = data;
    }

    public TopologyData getTopology() {
        return cachedTopology;
    }

    public static class ServiceNode {
        private final String name;
        private final List<ServiceInstance> instances = new ArrayList<>();
        private final AtomicLong incomingQps = new AtomicLong(0);
        private final AtomicLong outgoingQps = new AtomicLong(0);
        private final AtomicLong totalErrors = new AtomicLong(0);
        private final AtomicLong totalRequests = new AtomicLong(0);

        public ServiceNode(String name) {
            this.name = name;
        }

        public void addInstance(ServiceInstance inst) {
            instances.add(inst);
        }

        public String getName() { return name; }
        public int getInstanceCount() { return instances.size(); }
        public boolean isHealthy() {
            for (ServiceInstance inst : instances) {
                if (inst.isHealthy()) return true;
            }
            return !instances.isEmpty();
        }
        public long getTotalQps() { return incomingQps.get() + outgoingQps.get(); }
        public double getErrorRate() {
            long total = totalRequests.get();
            return total > 0 ? (double) totalErrors.get() / total * 100 : 0.0;
        }
    }

    public static class ServiceEdgeData {
        private final String source;
        private final String target;
        private double qps;
        private long totalRequests;
        private long errorCount;
        private double errorRate;

        public ServiceEdgeData(String source, String target) {
            this.source = source;
            this.target = target;
        }

        public void update(double qps, long totalRequests, long errorCount, double errorRate) {
            this.qps = qps;
            this.totalRequests += totalRequests;
            this.errorCount += errorCount;
            this.errorRate = this.totalRequests > 0 ? (double) this.errorCount / this.totalRequests * 100 : 0.0;
        }

        public String getSource() { return source; }
        public String getTarget() { return target; }
        public double getQps() { return qps; }
        public double getErrorRate() { return errorRate; }
    }
}
