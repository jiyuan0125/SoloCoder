package com.canary.router.core;

import com.canary.router.model.RoutingRule;
import com.canary.router.model.Stats;
import com.canary.router.model.Version;
import io.netty.handler.codec.http.*;

import java.net.InetSocketAddress;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ThreadLocalRandom;

public class RouterManager {
    private static final RouterManager INSTANCE = new RouterManager();
    
    private final Map<String, Version> versions = new ConcurrentHashMap<>();
    private final List<RoutingRule> rules = new ArrayList<>();
    private final Map<String, Stats> stats = new ConcurrentHashMap<>();
    private final Map<Integer, Integer> weightRangeCache = new HashMap<>();
    private int totalWeight = 0;
    private final Object lock = new Object();
    
    private RouterManager() {
        rules.add(new RoutingRule(RoutingRule.SourceType.HEADER, "canary", "true", "canary"));
        rules.add(new RoutingRule(RoutingRule.SourceType.COOKIE, "canary", "true", "canary"));
        rules.add(new RoutingRule(RoutingRule.SourceType.QUERY, "canary", "true", "canary"));
    }
    
    public static RouterManager getInstance() {
        return INSTANCE;
    }
    
    public void addVersion(Version version) {
        synchronized (lock) {
            versions.put(version.getName(), version);
            stats.computeIfAbsent(version.getName(), Stats::new);
            recalculateWeights();
        }
    }
    
    public boolean removeVersion(String name) {
        synchronized (lock) {
            boolean hasRules = rules.stream()
                .anyMatch(r -> r.getTargetVersion().equals(name));
            if (hasRules) {
                return false;
            }
            
            Version v = versions.remove(name);
            if (v != null) {
                v.setActive(false);
            }
            recalculateWeights();
            return true;
        }
    }
    
    public Version getVersion(String name) {
        return versions.get(name);
    }
    
    public Collection<Version> getActiveVersions() {
        return versions.values();
    }
    
    public void addRule(RoutingRule rule) {
        synchronized (lock) {
            rules.add(rule);
        }
    }
    
    public List<RoutingRule> getRules() {
        synchronized (lock) {
            return new ArrayList<>(rules);
        }
    }
    
    public Collection<Stats> getAllStats() {
        return stats.values();
    }
    
    public Stats getStats(String version) {
        return stats.computeIfAbsent(version, Stats::new);
    }
    
    public Version route(HttpRequest request, InetSocketAddress clientAddress) {
        String matchedVersion = matchExactRules(request, clientAddress);
        
        if (matchedVersion != null) {
            return versions.get(matchedVersion);
        }
        
        return routeByWeight();
    }
    
    private String matchExactRules(HttpRequest request, InetSocketAddress clientAddress) {
        String clientIp = clientAddress.getAddress().getHostAddress();
        
        Map<RoutingRule.SourceType, Integer> priority = new EnumMap<>(RoutingRule.SourceType.class);
        priority.put(RoutingRule.SourceType.HEADER, 4);
        priority.put(RoutingRule.SourceType.COOKIE, 3);
        priority.put(RoutingRule.SourceType.QUERY, 2);
        priority.put(RoutingRule.SourceType.IP, 1);
        
        RoutingRule bestMatch = null;
        int bestPriority = -1;
        
        for (RoutingRule rule : rules) {
            if (!versions.containsKey(rule.getTargetVersion())) {
                continue;
            }
            
            if (matchRule(rule, request, clientIp)) {
                int rulePriority = priority.getOrDefault(rule.getSource(), 0);
                if (rulePriority > bestPriority) {
                    bestPriority = rulePriority;
                    bestMatch = rule;
                }
            }
        }
        
        return bestMatch != null ? bestMatch.getTargetVersion() : null;
    }
    
    private boolean matchRule(RoutingRule rule, HttpRequest request, String clientIp) {
        switch (rule.getSource()) {
            case HEADER:
                String headerValue = request.headers().get(rule.getKey());
                return rule.getValue().equals(headerValue);
                
            case COOKIE:
                String cookieHeader = request.headers().get(HttpHeaderNames.COOKIE);
                if (cookieHeader != null) {
                    for (String cookiePart : cookieHeader.split(";")) {
                        cookiePart = cookiePart.trim();
                        int equals = cookiePart.indexOf('=');
                        if (equals > 0) {
                            String name = cookiePart.substring(0, equals);
                            String value = cookiePart.substring(equals + 1);
                            if (name.equals(rule.getKey()) && value.equals(rule.getValue())) {
                                return true;
                            }
                        }
                    }
                }
                return false;
                
            case QUERY:
                QueryStringDecoder queryDecoder = new QueryStringDecoder(request.uri());
                List<String> values = queryDecoder.parameters().get(rule.getKey());
                return values != null && values.contains(rule.getValue());
                
            case IP:
                return rule.getValue().equals(clientIp);
                
            default:
                return false;
        }
    }
    
    private Version routeByWeight() {
        if (totalWeight == 0) {
            return null;
        }
        
        int random = ThreadLocalRandom.current().nextInt(totalWeight);
        
        synchronized (lock) {
            for (Map.Entry<Integer, Integer> entry : weightRangeCache.entrySet()) {
                if (random < entry.getKey()) {
                    String versionName = getVersionNameByRangeIndex(entry.getValue());
                    return versions.get(versionName);
                }
            }
        }
        
        return null;
    }
    
    private void recalculateWeights() {
        synchronized (lock) {
            totalWeight = 0;
            weightRangeCache.clear();
            
            int index = 0;
            for (Version v : versions.values()) {
                if (v.getWeight() > 0) {
                    int start = totalWeight;
                    totalWeight += v.getWeight();
                    weightRangeCache.put(totalWeight, index);
                }
                index++;
            }
        }
    }
    
    private String getVersionNameByRangeIndex(int index) {
        int i = 0;
        for (String name : versions.keySet()) {
            if (i == index) {
                return name;
            }
            i++;
        }
        return null;
    }
}
