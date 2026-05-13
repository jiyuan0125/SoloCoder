package com.servicemesh.sidecar.service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.servicemesh.common.model.DiscoveryInstance;
import com.servicemesh.common.model.TrafficRule;
import com.servicemesh.sidecar.config.SidecarProperties;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicLong;

@Service
public class DiscoveryService {

    private static final Logger log = LoggerFactory.getLogger(DiscoveryService.class);

    private final SidecarProperties properties;
    private final OkHttpClient httpClient;
    private final ObjectMapper objectMapper;

    private final Map<String, List<DiscoveryInstance>> serviceInstances = new ConcurrentHashMap<>();
    private final Map<String, TrafficRule> trafficRules = new ConcurrentHashMap<>();

    private final Random random = new Random();
    private final Map<String, AtomicInteger> weightCounters = new ConcurrentHashMap<>();

    public DiscoveryService(SidecarProperties properties) {
        this.properties = properties;
        this.httpClient = new OkHttpClient.Builder()
                .connectTimeout(5, TimeUnit.SECONDS)
                .readTimeout(5, TimeUnit.SECONDS)
                .build();
        this.objectMapper = new ObjectMapper();
    }

    @PostConstruct
    public void init() {
        refreshRules();
    }

    public List<DiscoveryInstance> getInstances(String serviceName) {
        return serviceInstances.getOrDefault(serviceName, Collections.emptyList());
    }

    public DiscoveryInstance selectInstance(String serviceName) {
        List<DiscoveryInstance> instances = serviceInstances.get(serviceName);
        if (instances == null || instances.isEmpty()) {
            return null;
        }

        TrafficRule rule = trafficRules.get(serviceName);
        if (rule != null) {
            return applyRule(serviceName, instances, rule);
        }

        return instances.get(random.nextInt(instances.size()));
    }

    private DiscoveryInstance applyRule(String serviceName, List<DiscoveryInstance> instances, TrafficRule rule) {
        switch (rule.getType()) {
            case "VERSION_MATCH":
                return applyVersionMatch(instances, rule);
            case "WEIGHT_BASED":
                return applyWeightBased(serviceName, instances, rule);
            case "ZONE_BASED":
                return applyZoneBased(instances, rule);
            default:
                return instances.get(random.nextInt(instances.size()));
        }
    }

    private DiscoveryInstance applyVersionMatch(List<DiscoveryInstance> instances, TrafficRule rule) {
        String targetVersion = rule.getVersionMatch().getTargetVersion();
        List<DiscoveryInstance> matched = new ArrayList<>();
        for (DiscoveryInstance inst : instances) {
            if (targetVersion.equals(inst.getVersion())) {
                matched.add(inst);
            }
        }
        if (!matched.isEmpty()) {
            return matched.get(random.nextInt(matched.size()));
        }
        return instances.get(random.nextInt(instances.size()));
    }

    private DiscoveryInstance applyWeightBased(String serviceName, List<DiscoveryInstance> instances, TrafficRule rule) {
        Map<String, List<DiscoveryInstance>> byVersion = new HashMap<>();
        for (DiscoveryInstance inst : instances) {
            byVersion.computeIfAbsent(inst.getVersion(), k -> new ArrayList<>()).add(inst);
        }

        int totalWeight = 0;
        List<TrafficRule.VersionWeight> weights = rule.getWeightBased().getWeights();
        for (TrafficRule.VersionWeight w : weights) {
            if (byVersion.containsKey(w.getVersion())) {
                totalWeight += w.getWeight();
            }
        }

        AtomicInteger counter = weightCounters.computeIfAbsent(serviceName, k -> new AtomicInteger(0));
        int value = counter.getAndIncrement() % Math.max(totalWeight, 1);

        int cumulative = 0;
        for (TrafficRule.VersionWeight w : weights) {
            if (!byVersion.containsKey(w.getVersion())) continue;
            cumulative += w.getWeight();
            if (value < cumulative) {
                List<DiscoveryInstance> list = byVersion.get(w.getVersion());
                return list.get(random.nextInt(list.size()));
            }
        }

        return instances.get(random.nextInt(instances.size()));
    }

    private DiscoveryInstance applyZoneBased(List<DiscoveryInstance> instances, TrafficRule rule) {
        String preferZone = rule.getZoneBased().getPreferZone();
        List<DiscoveryInstance> sameZone = new ArrayList<>();
        List<DiscoveryInstance> others = new ArrayList<>();

        for (DiscoveryInstance inst : instances) {
            if (preferZone.equals(inst.getZone())) {
                sameZone.add(inst);
            } else {
                others.add(inst);
            }
        }

        if (!sameZone.isEmpty()) {
            return sameZone.get(random.nextInt(sameZone.size()));
        }

        if (rule.getZoneBased().isFallbackEnabled() && !others.isEmpty()) {
            return others.get(random.nextInt(others.size()));
        }

        return sameZone.isEmpty() ? (others.isEmpty() ? null : others.get(0)) : sameZone.get(0);
    }

    @Scheduled(fixedRateString = "${sidecar.rule-refresh-interval:5000}")
    public void refreshRules() {
        try {
            Request request = new Request.Builder()
                    .url(properties.getControlPlaneUrl() + "/api/rules")
                    .get()
                    .build();

            try (Response response = httpClient.newCall(request).execute()) {
                if (response.isSuccessful() && response.body() != null) {
                    JsonNode root = objectMapper.readTree(response.body().string());
                    trafficRules.clear();
                    for (JsonNode node : root) {
                        TrafficRule tr = objectMapper.treeToValue(node, TrafficRule.class);
                        trafficRules.put(tr.getServiceName(), tr);
                    }
                    log.debug("Refreshed {} traffic rules", trafficRules.size());
                }
            }
        } catch (Exception e) {
            log.debug("Rule refresh failed: {}", e.getMessage());
        }
    }

    public void updateServiceInstances(String serviceName, List<DiscoveryInstance> instances) {
        serviceInstances.put(serviceName, instances);
        log.info("Updated instances for {}: {} instances", serviceName, instances.size());
    }

    public void discoverService(String serviceName) {
        if (serviceInstances.containsKey(serviceName)) return;

        try {
            Request request = new Request.Builder()
                    .url(properties.getRegistryUrl() + "/api/registry/instances/" + serviceName + "?healthy=true")
                    .get()
                    .build();

            try (Response response = httpClient.newCall(request).execute()) {
                if (response.isSuccessful() && response.body() != null) {
                    JsonNode root = objectMapper.readTree(response.body().string());
                    List<DiscoveryInstance> instances = new ArrayList<>();
                    for (JsonNode node : root) {
                        DiscoveryInstance inst = objectMapper.treeToValue(node, DiscoveryInstance.class);
                        instances.add(inst);
                    }
                    serviceInstances.put(serviceName, instances);
                    log.info("Discovered {} instances for {}", instances.size(), serviceName);
                }
            }
        } catch (Exception e) {
            log.warn("Service discovery failed for {}: {}", serviceName, e.getMessage());
        }
    }

    public void subscribe(String serviceName, String callbackUrl) {
        try {
            Map<String, String> payload = Map.of(
                    "serviceName", serviceName,
                    "callbackUrl", callbackUrl
            );

            okhttp3.RequestBody body = okhttp3.RequestBody.create(
                    objectMapper.writeValueAsString(payload),
                    okhttp3.MediaType.parse("application/json; charset=utf-8")
            );
            Request request = new Request.Builder()
                    .url(properties.getRegistryUrl() + "/api/registry/subscribe")
                    .post(body)
                    .build();

            try (Response response = httpClient.newCall(request).execute()) {
                log.info("Subscribed to {}: {}", serviceName, response.isSuccessful());
            }
        } catch (Exception e) {
            log.warn("Subscribe failed", e);
        }
    }
}
