package com.trace.server.controller;

import com.trace.server.model.Span;
import com.trace.server.model.SpanTreeNode;
import com.trace.server.model.ServiceMapping;
import com.trace.server.model.ServiceTopology;
import com.trace.server.model.ServiceNode;
import com.trace.server.model.ServiceEdge;
import com.trace.server.service.SamplingService;
import com.trace.server.service.SpanStorageService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.*;

@RestController
@RequestMapping("/api/trace")
public class TraceController {

    private final SpanStorageService spanStorageService;
    private final SamplingService samplingService;

    public TraceController(SpanStorageService spanStorageService, SamplingService samplingService) {
        this.spanStorageService = spanStorageService;
        this.samplingService = samplingService;
    }

    @PostMapping("/span")
    public ResponseEntity<Map<String, String>> reportSpan(@RequestBody Span span) {
        Map<String, String> result = new HashMap<>();

        if (span.getSpanId() == null || span.getSpanId().isEmpty()) {
            span.setSpanId(UUID.randomUUID().toString().replace("-", "").substring(0, 16));
        }

        if (!samplingService.shouldSample(span)) {
            result.put("status", "not_sampled");
            result.put("message", "Span not sampled");
            return ResponseEntity.ok(result);
        }

        spanStorageService.saveSpan(span);

        result.put("status", "accepted");
        result.put("spanId", span.getSpanId());
        result.put("traceId", span.getTraceId());

        return ResponseEntity.ok(result);
    }

    @PostMapping("/spans/batch")
    public ResponseEntity<Map<String, Object>> reportSpans(@RequestBody List<Span> spans) {
        Map<String, Object> result = new HashMap<>();
        List<String> accepted = new ArrayList<>();
        List<String> notSampled = new ArrayList<>();

        for (Span span : spans) {
            if (span.getSpanId() == null || span.getSpanId().isEmpty()) {
                span.setSpanId(UUID.randomUUID().toString().replace("-", "").substring(0, 16));
            }

            if (samplingService.shouldSample(span)) {
                spanStorageService.saveSpan(span);
                accepted.add(span.getSpanId());
            } else {
                notSampled.add(span.getSpanId());
            }
        }

        result.put("accepted", accepted.size());
        result.put("notSampled", notSampled.size());
        result.put("acceptedSpanIds", accepted);

        return ResponseEntity.ok(result);
    }

    @GetMapping("/traces/{traceId}")
    public ResponseEntity<List<SpanTreeNode>> getTrace(@PathVariable String traceId) {
        List<SpanTreeNode> tree = spanStorageService.buildTraceTree(traceId);
        if (tree.isEmpty()) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(tree);
    }

    @GetMapping("/topology")
    public ResponseEntity<ServiceTopology> getServiceTopology() {
        Collection<Span> spans = spanStorageService.getAllSpans();
        Map<String, String> serviceNames = spanStorageService.getAllServiceNames();

        Set<String> serviceIds = new HashSet<>();
        Map<String, Long> callCounts = new HashMap<>();

        for (Span span : spans) {
            if (span.getServiceId() != null) {
                serviceIds.add(span.getServiceId());
            }
        }

        Map<String, Span> spanMap = new HashMap<>();
        for (Span span : spans) {
            spanMap.put(span.getSpanId(), span);
        }

        for (Span span : spans) {
            if (span.getParentSpanId() != null && !span.getParentSpanId().isEmpty()) {
                Span parent = spanMap.get(span.getParentSpanId());
                if (parent != null && parent.getServiceId() != null && span.getServiceId() != null) {
                    if (!parent.getServiceId().equals(span.getServiceId())) {
                        String key = parent.getServiceId() + "->" + span.getServiceId();
                        callCounts.merge(key, 1L, Long::sum);
                    }
                }
            }
        }

        List<ServiceNode> nodes = new ArrayList<>();
        for (String serviceId : serviceIds) {
            String name = serviceNames.getOrDefault(serviceId, serviceId);
            nodes.add(new ServiceNode(serviceId, name));
        }

        List<ServiceEdge> edges = new ArrayList<>();
        for (Map.Entry<String, Long> entry : callCounts.entrySet()) {
            String[] parts = entry.getKey().split("->");
            if (parts.length == 2) {
                edges.add(new ServiceEdge(parts[0], parts[1], entry.getValue()));
            }
        }

        ServiceTopology topology = new ServiceTopology();
        topology.setNodes(nodes);
        topology.setEdges(edges);

        return ResponseEntity.ok(topology);
    }

    @PutMapping("/services/{serviceId}/name")
    public ResponseEntity<ServiceMapping> updateServiceName(
            @PathVariable String serviceId,
            @RequestBody Map<String, String> body) {
        String newName = body.get("name");
        if (newName == null || newName.isEmpty()) {
            return ResponseEntity.badRequest().build();
        }

        spanStorageService.updateServiceName(serviceId, newName);
        ServiceMapping mapping = spanStorageService.getServiceMapping(serviceId);

        if (mapping == null) {
            return ResponseEntity.notFound().build();
        }

        return ResponseEntity.ok(mapping);
    }

    @GetMapping("/services/{serviceId}")
    public ResponseEntity<ServiceMapping> getServiceMapping(@PathVariable String serviceId) {
        ServiceMapping mapping = spanStorageService.getServiceMapping(serviceId);
        if (mapping == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(mapping);
    }

    @GetMapping("/services")
    public ResponseEntity<Map<String, String>> getAllServices() {
        Map<String, String> services = spanStorageService.getAllServiceNames();
        return ResponseEntity.ok(services);
    }
}
