package com.trace.collector.service;

import com.trace.collector.model.Span;
import com.trace.collector.model.SpanNode;
import com.trace.collector.model.TraceTree;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Service
public class TraceService {

    private final Map<String, Map<String, Span>> traceStore = new ConcurrentHashMap<>();
    private final Set<String> services = ConcurrentHashMap.newKeySet();
    private final Map<String, Set<String>> serviceOperations = new ConcurrentHashMap<>();

    public void saveSpans(List<Span> spans) {
        if (spans == null || spans.isEmpty()) {
            return;
        }

        for (Span span : spans) {
            String traceId = span.getTraceId();
            String spanId = span.getSpanId();

            traceStore.computeIfAbsent(traceId, k -> new ConcurrentHashMap<>()).put(spanId, span);

            String serviceName = span.getService();
            services.add(serviceName);

            serviceOperations
                    .computeIfAbsent(serviceName, k -> ConcurrentHashMap.newKeySet())
                    .add(span.getOperation());
        }
    }

    public List<String> getAllServices() {
        return new ArrayList<>(services).stream().sorted().collect(Collectors.toList());
    }

    public List<String> getOperationsByService(String serviceName) {
        Set<String> operations = serviceOperations.get(serviceName);
        if (operations == null) {
            return Collections.emptyList();
        }
        return new ArrayList<>(operations).stream().sorted().collect(Collectors.toList());
    }

    public TraceTree getTraceTree(String traceId) {
        Map<String, Span> spanMap = traceStore.get(traceId);
        if (spanMap == null || spanMap.isEmpty()) {
            return null;
        }

        Collection<Span> spans = spanMap.values();
        List<Span> rootSpans = spans.stream()
                .filter(s -> s.getParentSpanId() == null || s.getParentSpanId().isEmpty())
                .collect(Collectors.toList());

        if (rootSpans.isEmpty() && !spans.isEmpty()) {
            rootSpans.add(spans.iterator().next());
        }

        long totalDuration = rootSpans.stream().mapToLong(s -> s.getDuration() != null ? s.getDuration() : 0).sum();

        List<SpanNode> rootNodes = rootSpans.stream()
                .map(s -> buildNode(s, spanMap, totalDuration))
                .collect(Collectors.toList());

        TraceTree traceTree = new TraceTree();
        traceTree.setTraceId(traceId);
        traceTree.setSpans(rootNodes);

        return traceTree;
    }

    public List<TraceTree> searchTraces(String service, Instant from, Instant to) {
        List<TraceTree> results = new ArrayList<>();

        for (Map.Entry<String, Map<String, Span>> entry : traceStore.entrySet()) {
            String traceId = entry.getKey();
            Collection<Span> spans = entry.getValue().values();

            boolean matches = true;

            if (service != null && !service.isEmpty()) {
                boolean hasService = spans.stream().anyMatch(s -> service.equals(s.getService()));
                if (!hasService) {
                    matches = false;
                }
            }

            if (matches && from != null) {
                boolean afterFrom = spans.stream().anyMatch(s -> s.getStartTime() != null && !s.getStartTime().isBefore(from));
                if (!afterFrom) {
                    matches = false;
                }
            }

            if (matches && to != null) {
                boolean beforeTo = spans.stream().anyMatch(s -> s.getStartTime() != null && !s.getStartTime().isAfter(to));
                if (!beforeTo) {
                    matches = false;
                }
            }

            if (matches) {
                TraceTree traceTree = getTraceTree(traceId);
                if (traceTree != null) {
                    results.add(traceTree);
                }
            }
        }

        results.sort(Comparator.comparing((TraceTree t) -> {
            if (t.getSpans() == null || t.getSpans().isEmpty()) {
                return Instant.MIN;
            }
            Instant min = Instant.MAX;
            for (SpanNode node : t.getSpans()) {
                if (node.getStartTime() != null && node.getStartTime().isBefore(min)) {
                    min = node.getStartTime();
                }
            }
            return min;
        }).reversed());

        return results;
    }

    private SpanNode buildNode(Span span, Map<String, Span> allSpans, long totalDuration) {
        SpanNode node = new SpanNode();
        node.setTraceId(span.getTraceId());
        node.setSpanId(span.getSpanId());
        node.setParentSpanId(span.getParentSpanId());
        node.setService(span.getService());
        node.setOperation(span.getOperation());
        node.setStartTime(span.getStartTime());
        node.setDuration(span.getDuration());
        node.setStatus(span.getStatus());
        node.setTags(span.getTags());
        node.setSlow(span.isSlow());

        if (totalDuration > 0 && span.getDuration() != null) {
            node.setPercentage(Math.round((span.getDuration() * 10000.0) / totalDuration) / 100.0);
        } else {
            node.setPercentage(0.0);
        }

        List<SpanNode> children = allSpans.values().stream()
                .filter(s -> span.getSpanId().equals(s.getParentSpanId()))
                .map(s -> buildNode(s, allSpans, totalDuration))
                .sorted(Comparator.comparing(SpanNode::getStartTime, Comparator.nullsLast(Comparator.naturalOrder())))
                .collect(Collectors.toList());

        if (!children.isEmpty()) {
            node.setChildren(children);
        }

        return node;
    }
}
