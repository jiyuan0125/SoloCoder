package com.trace.server.service;

import com.trace.server.model.Span;
import com.trace.server.model.SpanTreeNode;
import com.trace.server.model.ServiceMapping;
import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.locks.ReentrantReadWriteLock;

@Service
public class SpanStorageService {

    private final Map<String, Span> spanStore = new ConcurrentHashMap<>();
    private final Map<String, List<Span>> traceIndex = new ConcurrentHashMap<>();
    private final Map<String, String> serviceNameMapping = new ConcurrentHashMap<>();
    private final Map<String, ServiceMapping> serviceMappingHistory = new ConcurrentHashMap<>();
    private final ReentrantReadWriteLock lock = new ReentrantReadWriteLock();

    public void saveSpan(Span span) {
        lock.writeLock().lock();
        try {
            if (span.getServiceName() != null && !span.getServiceName().isEmpty()) {
                if (span.getServiceId() == null || span.getServiceId().isEmpty()) {
                    span.setServiceId(span.getServiceName());
                }
                serviceNameMapping.put(span.getServiceId(), span.getServiceName());
            }
            spanStore.put(span.getSpanId(), span);
            traceIndex.computeIfAbsent(span.getTraceId(), k -> new ArrayList<>()).add(span);
        } finally {
            lock.writeLock().unlock();
        }
    }

    public List<Span> getSpansByTraceId(String traceId) {
        lock.readLock().lock();
        try {
            List<Span> spans = traceIndex.getOrDefault(traceId, Collections.emptyList());
            spans.forEach(this::applyServiceNameMapping);
            return spans;
        } finally {
            lock.readLock().unlock();
        }
    }

    public List<SpanTreeNode> buildTraceTree(String traceId) {
        List<Span> spans = getSpansByTraceId(traceId);
        if (spans.isEmpty()) {
            return Collections.emptyList();
        }

        Map<String, SpanTreeNode> nodeMap = new HashMap<>();
        Map<String, List<String>> parentChildrenMap = new HashMap<>();

        long maxDuration = 0;
        String bottleneckSpanId = null;

        for (Span span : spans) {
            SpanTreeNode node = new SpanTreeNode();
            node.setSpan(span);
            long duration = span.getEndTime() != null && span.getStartTime() != null 
                ? span.getEndTime() - span.getStartTime() : 0;
            node.setDuration(duration);
            node.setIsBottleneck(false);
            node.setChildren(new ArrayList<>());
            nodeMap.put(span.getSpanId(), node);

            if (duration > maxDuration) {
                maxDuration = duration;
                bottleneckSpanId = span.getSpanId();
            }

            if (span.getParentSpanId() != null && !span.getParentSpanId().isEmpty()) {
                parentChildrenMap.computeIfAbsent(span.getParentSpanId(), k -> new ArrayList<>())
                    .add(span.getSpanId());
            }
        }

        if (bottleneckSpanId != null && nodeMap.containsKey(bottleneckSpanId)) {
            nodeMap.get(bottleneckSpanId).setIsBottleneck(true);
        }

        List<SpanTreeNode> roots = new ArrayList<>();
        for (Span span : spans) {
            SpanTreeNode node = nodeMap.get(span.getSpanId());
            String parentId = span.getParentSpanId();
            if (parentId == null || parentId.isEmpty() || !nodeMap.containsKey(parentId)) {
                roots.add(node);
            } else {
                SpanTreeNode parent = nodeMap.get(parentId);
                if (parent.getChildren() == null) {
                    parent.setChildren(new ArrayList<>());
                }
                parent.getChildren().add(node);
            }
        }

        roots.sort(Comparator.comparingLong(n -> 
            n.getSpan().getStartTime() != null ? n.getSpan().getStartTime() : 0));
        return roots;
    }

    public void updateServiceName(String serviceId, String newName) {
        lock.writeLock().lock();
        try {
            String oldName = serviceNameMapping.get(serviceId);
            if (oldName != null && !oldName.equals(newName)) {
                ServiceMapping mapping = new ServiceMapping();
                mapping.setServiceId(serviceId);
                mapping.setPreviousName(oldName);
                mapping.setCurrentName(newName);
                mapping.setUpdatedAt(Instant.now());
                serviceMappingHistory.put(serviceId, mapping);
            }
            serviceNameMapping.put(serviceId, newName);
        } finally {
            lock.writeLock().unlock();
        }
    }

    public ServiceMapping getServiceMapping(String serviceId) {
        lock.readLock().lock();
        try {
            ServiceMapping history = serviceMappingHistory.get(serviceId);
            if (history == null) {
                String currentName = serviceNameMapping.get(serviceId);
                if (currentName != null) {
                    ServiceMapping m = new ServiceMapping();
                    m.setServiceId(serviceId);
                    m.setCurrentName(currentName);
                    return m;
                }
                return null;
            }
            return history;
        } finally {
            lock.readLock().unlock();
        }
    }

    public Map<String, String> getAllServiceNames() {
        lock.readLock().lock();
        try {
            return new HashMap<>(serviceNameMapping);
        } finally {
            lock.readLock().unlock();
        }
    }

    public Collection<Span> getAllSpans() {
        lock.readLock().lock();
        try {
            ArrayList<Span> spans = new ArrayList<>(spanStore.values());
            spans.forEach(this::applyServiceNameMapping);
            return spans;
        } finally {
            lock.readLock().unlock();
        }
    }

    private void applyServiceNameMapping(Span span) {
        if (span.getArchived() != null && span.getArchived()) {
            return;
        }
        String mappedName = serviceNameMapping.get(span.getServiceId());
        if (mappedName != null) {
            span.setServiceName(mappedName);
        }
    }
}
