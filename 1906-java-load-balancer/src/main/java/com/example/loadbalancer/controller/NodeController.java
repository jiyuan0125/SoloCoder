package com.example.loadbalancer.controller;

import com.example.loadbalancer.dto.NodeResponse;
import com.example.loadbalancer.dto.NodesListResponse;
import com.example.loadbalancer.dto.RegisterNodeRequest;
import com.example.loadbalancer.model.Node;
import com.example.loadbalancer.registry.NodeRegistry;
import com.example.loadbalancer.strategy.StrategyManager;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/nodes")
public class NodeController {

    private final NodeRegistry nodeRegistry;
    private final StrategyManager strategyManager;

    public NodeController(NodeRegistry nodeRegistry, StrategyManager strategyManager) {
        this.nodeRegistry = nodeRegistry;
        this.strategyManager = strategyManager;
    }

    @PostMapping
    public ResponseEntity<?> registerNode(@Valid @RequestBody RegisterNodeRequest request) {
        Optional<Node> node = nodeRegistry.registerNode(
                request.getIp(),
                request.getPort(),
                request.getWeight()
        );

        if (node.isEmpty()) {
            return ResponseEntity.status(HttpStatus.CONFLICT)
                    .body(Map.of("error", "Node with address " + request.getIp() + ":" + request.getPort() + " already exists"));
        }

        return ResponseEntity.status(HttpStatus.CREATED).body(toNodeResponse(node.get()));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<?> offlineNode(@PathVariable String id) {
        Optional<Node> node = nodeRegistry.offlineNode(id);
        
        if (node.isEmpty()) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(Map.of("error", "Node with id " + id + " not found"));
        }

        return ResponseEntity.ok(toNodeResponse(node.get()));
    }

    @GetMapping
    public ResponseEntity<NodesListResponse> getNodes() {
        List<NodeResponse> nodeResponses = nodeRegistry.getAllNodes().stream()
                .map(this::toNodeResponse)
                .collect(Collectors.toList());

        NodesListResponse response = NodesListResponse.builder()
                .currentStrategy(strategyManager.getCurrentStrategyName())
                .nodes(nodeResponses)
                .build();

        return ResponseEntity.ok(response);
    }

    private NodeResponse toNodeResponse(Node node) {
        return NodeResponse.builder()
                .id(node.getId())
                .ip(node.getIp())
                .port(node.getPort())
                .initialWeight(node.getInitialWeight())
                .currentWeight(node.getCurrentWeight())
                .status(node.getStatus().name())
                .activeConnections(node.getActiveConnectionsCount())
                .registeredAt(node.getRegisteredAt())
                .build();
    }
}
