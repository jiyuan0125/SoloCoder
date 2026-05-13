package com.loadbalancer.api;

import com.loadbalancer.dto.ApiResponse;
import com.loadbalancer.dto.RegisterNodeRequest;
import com.loadbalancer.manager.NodeManager;
import com.loadbalancer.model.Node;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.Comparator;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/nodes")
public class NodeController {

    private final NodeManager nodeManager;

    @Autowired
    public NodeController(NodeManager nodeManager) {
        this.nodeManager = nodeManager;
    }

    @PostMapping
    public ResponseEntity<ApiResponse<Node>> registerNode(@Valid @RequestBody RegisterNodeRequest request) {
        Optional<Node> nodeOpt = nodeManager.registerNode(request.getIp(), request.getPort());
        if (nodeOpt.isPresent()) {
            return ResponseEntity.status(HttpStatus.CREATED)
                    .body(ApiResponse.success("Node registered", nodeOpt.get()));
        }
        return ResponseEntity.status(HttpStatus.CONFLICT)
                .body(ApiResponse.error("Node already exists: " + request.getIp() + ":" + request.getPort()));
    }

    @GetMapping
    public ResponseEntity<ApiResponse<List<Node>>> listNodes() {
        List<Node> nodes = nodeManager.getAllNodes().stream()
                .sorted(Comparator.comparing(Node::getKey))
                .collect(Collectors.toList());
        return ResponseEntity.ok(ApiResponse.success(nodes));
    }

    @GetMapping("/{ip}:{port}")
    public ResponseEntity<ApiResponse<Node>> getNode(@PathVariable String ip, @PathVariable int port) {
        Optional<Node> nodeOpt = nodeManager.getNode(ip, port);
        if (nodeOpt.isPresent()) {
            return ResponseEntity.ok(ApiResponse.success(nodeOpt.get()));
        }
        return ResponseEntity.status(HttpStatus.NOT_FOUND)
                .body(ApiResponse.error("Node not found"));
    }

    @PostMapping("/{ip}:{port}/activate")
    public ResponseEntity<ApiResponse<Node>> activateNode(@PathVariable String ip, @PathVariable int port) {
        boolean activated = nodeManager.activateNode(ip, port);
        if (activated) {
            Optional<Node> node = nodeManager.getNode(ip, port);
            return ResponseEntity.ok(ApiResponse.success("Node activated", node.orElse(null)));
        }
        return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                .body(ApiResponse.error("Cannot activate node - may not exist or not in REGISTERING state"));
    }

    @PostMapping("/{ip}:{port}/drain")
    public ResponseEntity<ApiResponse<Node>> startDraining(@PathVariable String ip, @PathVariable int port) {
        boolean drained = nodeManager.startDraining(ip, port);
        if (drained) {
            Optional<Node> node = nodeManager.getNode(ip, port);
            return ResponseEntity.ok(ApiResponse.success("Node started draining", node.orElse(null)));
        }
        return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                .body(ApiResponse.error("Cannot drain node - may not exist or not in ACTIVE state"));
    }

    @PostMapping("/{ip}:{port}/offline")
    public ResponseEntity<ApiResponse<Node>> markOffline(@PathVariable String ip, @PathVariable int port) {
        boolean offline = nodeManager.markOffline(ip, port);
        if (offline) {
            Optional<Node> node = nodeManager.getNode(ip, port);
            return ResponseEntity.ok(ApiResponse.success("Node marked offline", node.orElse(null)));
        }
        return ResponseEntity.status(HttpStatus.NOT_FOUND)
                .body(ApiResponse.error("Node not found"));
    }

    @DeleteMapping("/{ip}:{port}")
    public ResponseEntity<ApiResponse<Void>> removeNode(@PathVariable String ip, @PathVariable int port) {
        boolean removed = nodeManager.removeNode(ip, port);
        if (removed) {
            return ResponseEntity.ok(ApiResponse.success("Node removed", null));
        }
        return ResponseEntity.status(HttpStatus.NOT_FOUND)
                .body(ApiResponse.error("Node not found"));
    }
}
