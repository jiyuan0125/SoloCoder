package com.trace.collector.store;

import com.trace.collector.model.Span;
import com.trace.collector.model.TraceNode;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.stream.Collectors;

public class SpanStore {

    private static final Logger logger = LoggerFactory.getLogger(SpanStore.class);

    private final Map<String, List<Span>> traceMap = new ConcurrentHashMap<>();
    private final List<Span> allSpans = new CopyOnWriteArrayList<>();

    public void addSpan(Span span) {
        traceMap.computeIfAbsent(span.getTraceId(), k -> new CopyOnWriteArrayList<>())
                .add(span);
        allSpans.add(span);
    }

    public List<TraceNode> getTraceTree(String traceId) {
        List<Span> spans = traceMap.get(traceId);
        if (spans == null || spans.isEmpty()) {
            return Collections.emptyList();
        }
        return buildTraceTree(spans);
    }

    public List<Span> searchSpans(String serviceName, Long from, Long to) {
        return allSpans.stream()
                .filter(span -> serviceName == null || serviceName.isEmpty() ||
                        serviceName.equals(span.getServiceName()))
                .filter(span -> {
                    if (from == null && to == null) return true;
                    Long ts = span.getStartTimestamp();
                    if (ts == null) return false;
                    if (from != null && ts < from) return false;
                    if (to != null && ts > to) return false;
                    return true;
                })
                .sorted(Comparator.comparingLong((Span s) ->
                        s.getStartTimestamp() != null ? s.getStartTimestamp() : 0L
                ).reversed())
                .collect(Collectors.toList());
    }

    private List<TraceNode> buildTraceTree(List<Span> spans) {
        Map<String, TraceNode> nodeMap = new HashMap<>();
        List<TraceNode> roots = new ArrayList<>();

        for (Span span : spans) {
            nodeMap.put(span.getSpanId(), new TraceNode(span));
        }

        for (Span span : spans) {
            TraceNode node = nodeMap.get(span.getSpanId());
            if (span.getParentSpanId() == null || span.getParentSpanId().isEmpty()) {
                roots.add(node);
            } else {
                TraceNode parent = nodeMap.get(span.getParentSpanId());
                if (parent != null) {
                    parent.addChild(node);
                } else {
                    roots.add(node);
                }
            }
        }

        return roots;
    }
}
