package com.depgraph.service;

import com.depgraph.model.CycleAlert;
import com.depgraph.model.DependencyEdge;
import com.depgraph.model.GraphData;
import com.depgraph.model.ImpactResult;
import com.depgraph.model.ServiceNode;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.Queue;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Slf4j
@Service
public class DependencyGraphService {

    private final Map<String, ServiceNode> nodes = new ConcurrentHashMap<>();
    private final Map<String, Set<String>> adjacency = new ConcurrentHashMap<>();
    private final Map<String, Set<String>> reverseAdjacency = new ConcurrentHashMap<>();
    private final Map<String, DependencyEdge> edges = new ConcurrentHashMap<>();

    public synchronized CycleAlert registerService(String serviceName, List<String> dependencies) {
        log.info("Registering service: {} with dependencies: {}", serviceName, dependencies);

        nodes.putIfAbsent(serviceName, new ServiceNode(serviceName));
        adjacency.putIfAbsent(serviceName, new HashSet<>());
        reverseAdjacency.putIfAbsent(serviceName, new HashSet<>());

        removeOldDependencies(serviceName);

        if (dependencies != null) {
            for (String dep : dependencies) {
                if (dep == null || dep.isEmpty() || dep.equals(serviceName)) {
                    continue;
                }
                addDependency(serviceName, dep);
            }
        }

        return detectCycle();
    }

    private void removeOldDependencies(String serviceName) {
        Set<String> oldDeps = adjacency.getOrDefault(serviceName, Collections.emptySet());
        for (String dep : new ArrayList<>(oldDeps)) {
            removeDependency(serviceName, dep);
        }
    }

    private void addDependency(String from, String to) {
        nodes.putIfAbsent(to, new ServiceNode(to));
        adjacency.putIfAbsent(to, new HashSet<>());
        reverseAdjacency.putIfAbsent(to, new HashSet<>());

        adjacency.get(from).add(to);
        reverseAdjacency.get(to).add(from);

        String edgeKey = edgeKey(from, to);
        edges.computeIfAbsent(edgeKey, k -> new DependencyEdge(from, to));
    }

    private void removeDependency(String from, String to) {
        if (adjacency.containsKey(from)) {
            adjacency.get(from).remove(to);
        }
        if (reverseAdjacency.containsKey(to)) {
            reverseAdjacency.get(to).remove(from);
        }
        edges.remove(edgeKey(from, to));
    }

    private String edgeKey(String from, String to) {
        return from + "->" + to;
    }

    public CycleAlert detectCycle() {
        Set<String> visited = new HashSet<>();
        Set<String> recursionStack = new HashSet<>();
        Map<String, String> parent = new HashMap<>();

        for (String node : nodes.keySet()) {
            if (!visited.contains(node)) {
                List<String> cycle = dfsCycle(node, visited, recursionStack, parent);
                if (cycle != null) {
                    log.warn("Cycle detected: {}", cycle);
                    return CycleAlert.found(cycle);
                }
            }
        }

        return CycleAlert.none();
    }

    private List<String> dfsCycle(String current, Set<String> visited, Set<String> recursionStack,
                                   Map<String, String> parent) {
        visited.add(current);
        recursionStack.add(current);

        for (String neighbor : adjacency.getOrDefault(current, Collections.emptySet())) {
            parent.put(neighbor, current);

            if (!visited.contains(neighbor)) {
                List<String> cycle = dfsCycle(neighbor, visited, recursionStack, parent);
                if (cycle != null) {
                    return cycle;
                }
            } else if (recursionStack.contains(neighbor)) {
                return buildCyclePath(parent, neighbor, current);
            }
        }

        recursionStack.remove(current);
        return null;
    }

    private List<String> buildCyclePath(Map<String, String> parent, String start, String end) {
        List<String> cycle = new ArrayList<>();
        cycle.add(start);

        String current = end;
        while (!current.equals(start)) {
            cycle.add(current);
            current = parent.get(current);
        }
        cycle.add(start);
        Collections.reverse(cycle);
        return cycle;
    }

    public ImpactResult analyzeImpact(String serviceName) {
        if (!nodes.containsKey(serviceName)) {
            return new ImpactResult(serviceName, Collections.emptyList());
        }

        List<ImpactResult.AffectedService> affected = new ArrayList<>();
        Map<String, Integer> depthMap = new HashMap<>();
        Map<String, String> parentMap = new HashMap<>();
        Set<String> visited = new HashSet<>();
        Queue<String> queue = new LinkedList<>();

        depthMap.put(serviceName, 0);
        queue.offer(serviceName);
        visited.add(serviceName);

        while (!queue.isEmpty()) {
            String current = queue.poll();

            for (String downstream : reverseAdjacency.getOrDefault(current, Collections.emptySet())) {
                if (!visited.contains(downstream)) {
                    visited.add(downstream);
                    parentMap.put(downstream, current);
                    depthMap.put(downstream, depthMap.get(current) + 1);
                    queue.offer(downstream);

                    boolean direct = current.equals(serviceName);
                    affected.add(new ImpactResult.AffectedService(
                            downstream,
                            depthMap.get(downstream),
                            buildPath(parentMap, serviceName, downstream),
                            direct
                    ));
                }
            }
        }

        affected.sort((a, b) -> Integer.compare(a.getDependencyDepth(), b.getDependencyDepth()));
        return new ImpactResult(serviceName, affected);
    }

    private List<String> buildPath(Map<String, String> parentMap, String start, String end) {
        List<String> path = new ArrayList<>();
        String current = end;
        while (current != null) {
            path.add(current);
            current = parentMap.get(current);
        }
        Collections.reverse(path);
        return path;
    }

    public GraphData getGraphData() {
        List<GraphData.NodeInfo> nodeList = nodes.values().stream()
                .map(n -> new GraphData.NodeInfo(n.getName(), n.getStatus()))
                .collect(Collectors.toList());

        List<GraphData.EdgeInfo> edgeList = edges.values().stream()
                .map(e -> new GraphData.EdgeInfo(e.getFrom(), e.getTo(), e.getCallCount()))
                .collect(Collectors.toList());

        return new GraphData(nodeList, edgeList);
    }

    public boolean updateHealth(String serviceName, ServiceNode.HealthStatus status) {
        ServiceNode node = nodes.get(serviceName);
        if (node != null) {
            node.setStatus(status);
            node.setLastHeartbeat(LocalDateTime.now());
            log.info("Health updated for {}: {}", serviceName, status);
            return true;
        }
        return false;
    }

    public boolean heartbeat(String serviceName) {
        return updateHealth(serviceName, ServiceNode.HealthStatus.HEALTHY);
    }

    public ServiceNode getService(String serviceName) {
        return nodes.get(serviceName);
    }

    public List<ServiceNode> getAllServices() {
        return new ArrayList<>(nodes.values());
    }

    public boolean unregisterService(String serviceName) {
        if (!nodes.containsKey(serviceName)) {
            return false;
        }

        for (String dep : new ArrayList<>(adjacency.getOrDefault(serviceName, Collections.emptySet()))) {
            removeDependency(serviceName, dep);
        }

        for (String dependent : new ArrayList<>(reverseAdjacency.getOrDefault(serviceName, Collections.emptySet()))) {
            removeDependency(dependent, serviceName);
        }

        nodes.remove(serviceName);
        adjacency.remove(serviceName);
        reverseAdjacency.remove(serviceName);

        log.info("Service unregistered: {}", serviceName);
        return true;
    }
}
