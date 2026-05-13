package com.loadbalancer.manager;

import com.loadbalancer.model.Node;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.Instant;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Component
public class SessionMigrationManager {

    private final Map<String, SessionMapping> sessionMappings = new ConcurrentHashMap<>();
    private final NodeManager nodeManager;

    @Value("${loadbalancer.session.window-seconds:60}")
    private long windowSeconds;

    public SessionMigrationManager(NodeManager nodeManager) {
        this.nodeManager = nodeManager;
    }

    public void setWindowSeconds(long windowSeconds) {
        this.windowSeconds = windowSeconds;
    }

    public long getWindowSeconds() {
        return windowSeconds;
    }

    public Node route(String sessionId) {
        Node target;

        SessionMapping existing = sessionMappings.get(sessionId);
        if (existing != null) {
            Instant now = Instant.now();
            boolean withinWindow = existing.getMappedAt().plusSeconds(windowSeconds).isAfter(now);
            Node mappedNode = nodeManager.getNodeForExistingSession(existing.getNodeKey());
            if (withinWindow && mappedNode != null) {
                return mappedNode;
            }
            sessionMappings.remove(sessionId);
        }

        if (nodeManager.isEmpty()) {
            return null;
        }

        target = nodeManager.getTargetNode(sessionId);
        if (target != null) {
            sessionMappings.put(sessionId, new SessionMapping(target.getKey(), Instant.now()));
        }
        return target;
    }

    public void recordMapping(String sessionId, String nodeKey) {
        sessionMappings.put(sessionId, new SessionMapping(nodeKey, Instant.now()));
    }

    @Scheduled(fixedRate = 10000)
    public void cleanExpiredMappings() {
        Instant cutoff = Instant.now().minusSeconds(windowSeconds);
        sessionMappings.entrySet().removeIf(entry ->
                entry.getValue().getMappedAt().isBefore(cutoff));
    }

    public int getMappingCount() {
        return sessionMappings.size();
    }

    public static class SessionMapping {
        private final String nodeKey;
        private final Instant mappedAt;

        public SessionMapping(String nodeKey, Instant mappedAt) {
            this.nodeKey = nodeKey;
            this.mappedAt = mappedAt;
        }

        public String getNodeKey() {
            return nodeKey;
        }

        public Instant getMappedAt() {
            return mappedAt;
        }
    }
}
