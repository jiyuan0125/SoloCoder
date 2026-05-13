package com.loadbalancer.strategy;

import com.loadbalancer.model.Node;
import org.springframework.stereotype.Component;

import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.Collection;
import java.util.Collections;
import java.util.Map;
import java.util.TreeMap;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.locks.ReentrantReadWriteLock;

@Component
public class ConsistentHashRing {

    private final TreeMap<Long, String> ring = new TreeMap<>();
    private final Map<String, Node> nodeMap = new ConcurrentHashMap<>();
    private final ReentrantReadWriteLock lock = new ReentrantReadWriteLock();

    private int vnodesPerNode = 150;

    public void setVnodesPerNode(int vnodesPerNode) {
        lock.writeLock().lock();
        try {
            this.vnodesPerNode = vnodesPerNode;
            rebuildRing();
        } finally {
            lock.writeLock().unlock();
        }
    }

    public int getVnodesPerNode() {
        return vnodesPerNode;
    }

    public void addNode(Node node) {
        lock.writeLock().lock();
        try {
            nodeMap.put(node.getKey(), node);
            for (int i = 0; i < vnodesPerNode; i++) {
                long hash = hash(node.getKey() + "#" + i);
                ring.put(hash, node.getKey());
            }
        } finally {
            lock.writeLock().unlock();
        }
    }

    public void removeNode(String nodeKey) {
        lock.writeLock().lock();
        try {
            nodeMap.remove(nodeKey);
            ring.entrySet().removeIf(entry -> entry.getValue().equals(nodeKey));
        } finally {
            lock.writeLock().unlock();
        }
    }

    public Node getNodeForKey(String key) {
        lock.readLock().lock();
        try {
            if (ring.isEmpty()) {
                return null;
            }
            long hash = hash(key);
            Map.Entry<Long, String> entry = ring.ceilingEntry(hash);
            if (entry == null) {
                entry = ring.firstEntry();
            }
            return nodeMap.get(entry.getValue());
        } finally {
            lock.readLock().unlock();
        }
    }

    public boolean isEmpty() {
        lock.readLock().lock();
        try {
            return nodeMap.isEmpty();
        } finally {
            lock.readLock().unlock();
        }
    }

    public Collection<Node> getAllNodes() {
        return Collections.unmodifiableCollection(nodeMap.values());
    }

    public Node getNode(String nodeKey) {
        return nodeMap.get(nodeKey);
    }

    public boolean containsNode(String nodeKey) {
        return nodeMap.containsKey(nodeKey);
    }

    private void rebuildRing() {
        ring.clear();
        for (Node node : nodeMap.values()) {
            for (int i = 0; i < vnodesPerNode; i++) {
                long hash = hash(node.getKey() + "#" + i);
                ring.put(hash, node.getKey());
            }
        }
    }

    private long hash(String key) {
        try {
            MessageDigest md = MessageDigest.getInstance("MD5");
            byte[] digest = md.digest(key.getBytes(StandardCharsets.UTF_8));
            long h1 = ((long)(digest[0] & 0xFF))
                    | (((long)(digest[1] & 0xFF)) << 8)
                    | (((long)(digest[2] & 0xFF)) << 16)
                    | (((long)(digest[3] & 0xFF)) << 24);
            long h2 = ((long)(digest[4] & 0xFF))
                    | (((long)(digest[5] & 0xFF)) << 8)
                    | (((long)(digest[6] & 0xFF)) << 16)
                    | (((long)(digest[7] & 0xFF)) << 24);
            return (h1 << 32) | (h2 & 0xFFFFFFFFL);
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException("MD5 not available", e);
        }
    }
}
