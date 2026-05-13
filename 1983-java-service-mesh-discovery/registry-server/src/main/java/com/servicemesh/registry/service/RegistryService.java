package com.servicemesh.registry.service;

import com.servicemesh.common.model.ServiceInstance;
import com.servicemesh.registry.config.RegistryProperties;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArraySet;

@Service
public class RegistryService {

    private static final Logger log = LoggerFactory.getLogger(RegistryService.class);

    private final Map<String, Map<String, ServiceInstance>> registry = new ConcurrentHashMap<>();
    private final Map<String, Set<String>> subscribers = new ConcurrentHashMap<>();
    private final Map<String, List<ServiceInstance>> changeNotifications = new ConcurrentHashMap<>();

    private final RegistryProperties properties;

    public RegistryService(RegistryProperties properties) {
        this.properties = properties;
    }

    public String register(ServiceInstance instance) {
        String instanceId = generateInstanceId(instance);
        instance.setInstanceId(instanceId);

        registry.computeIfAbsent(instance.getServiceName(), k -> new ConcurrentHashMap<>())
                .put(instanceId, instance);

        subscribers.computeIfAbsent(instance.getServiceName(), k -> new CopyOnWriteArraySet<>());

        log.info("Service registered: {}@{}:{} (version={}, zone={}, id={})",
                instance.getServiceName(), instance.getIp(), instance.getPort(),
                instance.getVersion(), instance.getZone(), instanceId);

        notifySubscribers(instance.getServiceName());

        return instanceId;
    }

    public boolean heartbeat(String serviceName, String instanceId) {
        Map<String, ServiceInstance> instances = registry.get(serviceName);
        if (instances == null) return false;

        ServiceInstance instance = instances.get(instanceId);
        if (instance == null) return false;

        instance.setLastHeartbeat(System.currentTimeMillis());
        instance.setMissedHeartbeats(0);

        if (!instance.isHealthy()) {
            if (doHealthCheck(instance)) {
                instance.setHealthy(true);
                log.info("Instance became healthy: {}@{}:{}", instance.getServiceName(), instance.getIp(), instance.getPort());
                notifySubscribers(serviceName);
            }
        }

        return true;
    }

    public void deregister(String serviceName, String instanceId) {
        Map<String, ServiceInstance> instances = registry.get(serviceName);
        if (instances != null) {
            ServiceInstance removed = instances.remove(instanceId);
            if (removed != null) {
                log.info("Service deregistered: {}@{}:{}", serviceName, removed.getIp(), removed.getPort());
                notifySubscribers(serviceName);
            }
        }
    }

    public List<ServiceInstance> getInstances(String serviceName, boolean onlyHealthy) {
        Map<String, ServiceInstance> instances = registry.get(serviceName);
        if (instances == null || instances.isEmpty()) {
            return new ArrayList<>();
        }

        List<ServiceInstance> result = new ArrayList<>();
        for (ServiceInstance instance : instances.values()) {
            if (!onlyHealthy || instance.isHealthy()) {
                result.add(instance);
            }
        }
        return result;
    }

    public List<ServiceInstance> getAllInstances(boolean onlyHealthy) {
        List<ServiceInstance> result = new ArrayList<>();
        for (Map<String, ServiceInstance> instances : registry.values()) {
            for (ServiceInstance instance : instances.values()) {
                if (!onlyHealthy || instance.isHealthy()) {
                    result.add(instance);
                }
            }
        }
        return result;
    }

    public List<String> getAllServiceNames() {
        return new ArrayList<>(registry.keySet());
    }

    public void subscribe(String serviceName, String callbackUrl) {
        subscribers.computeIfAbsent(serviceName, k -> new CopyOnWriteArraySet<>()).add(callbackUrl);
        log.info("New subscriber for {}: {}", serviceName, callbackUrl);
    }

    public void unsubscribe(String serviceName, String callbackUrl) {
        Set<String> urls = subscribers.get(serviceName);
        if (urls != null) {
            urls.remove(callbackUrl);
        }
    }

    @Scheduled(fixedRate = 5000)
    public void checkHeartbeats() {
        long now = System.currentTimeMillis();
        long timeout = properties.getHeartbeatInterval() * 2;

        for (Map.Entry<String, Map<String, ServiceInstance>> serviceEntry : registry.entrySet()) {
            String serviceName = serviceEntry.getKey();
            Map<String, ServiceInstance> instances = serviceEntry.getValue();
            boolean changed = false;

            for (ServiceInstance instance : instances.values()) {
                long elapsed = now - instance.getLastHeartbeat();

                if (elapsed > timeout) {
                    instance.setMissedHeartbeats(instance.getMissedHeartbeats() + 1);
                    log.debug("Instance heartbeat missed: {}@{}:{} (count={})",
                            instance.getServiceName(), instance.getIp(), instance.getPort(), instance.getMissedHeartbeats());

                    if (instance.getMissedHeartbeats() >= properties.getMaxMissedHeartbeats()) {
                        if (instance.isHealthy()) {
                            instance.setHealthy(false);
                            changed = true;
                            log.warn("Instance marked unhealthy: {}@{}:{}", instance.getServiceName(), instance.getIp(), instance.getPort());
                        }
                    }
                } else if (!instance.isHealthy() && elapsed <= timeout) {
                    if (doHealthCheck(instance)) {
                        instance.setHealthy(true);
                        instance.setMissedHeartbeats(0);
                        changed = true;
                        log.info("Instance recovered: {}@{}:{}", instance.getServiceName(), instance.getIp(), instance.getPort());
                    }
                }
            }

            if (changed) {
                notifySubscribers(serviceName);
            }
        }
    }

    private boolean doHealthCheck(ServiceInstance instance) {
        String url = "http://" + instance.getIp() + ":" + instance.getPort() + "/actuator/health";
        try {
            HttpURLConnection conn = (HttpURLConnection) new URL(url).openConnection();
            conn.setRequestMethod("GET");
            conn.setConnectTimeout(properties.getHealthCheckTimeout());
            conn.setReadTimeout(properties.getHealthCheckTimeout());

            int code = conn.getResponseCode();
            if (code == 200) {
                BufferedReader reader = new BufferedReader(new InputStreamReader(conn.getInputStream()));
                StringBuilder body = new StringBuilder();
                String line;
                while ((line = reader.readLine()) != null) {
                    body.append(line);
                }
                reader.close();
                return body.toString().contains("\"status\":\"UP\"");
            }
            return false;
        } catch (Exception e) {
            log.debug("Health check failed for {}:{} - {}", instance.getIp(), instance.getPort(), e.getMessage());
            return false;
        }
    }

    private void notifySubscribers(String serviceName) {
        Set<String> urls = subscribers.get(serviceName);
        if (urls == null || urls.isEmpty()) return;

        List<ServiceInstance> instances = getInstances(serviceName, true);

        for (String callbackUrl : urls) {
            try {
                sendNotification(callbackUrl, serviceName, instances);
            } catch (Exception e) {
                log.warn("Failed to notify subscriber {}: {}", callbackUrl, e.getMessage());
            }
        }
    }

    private void sendNotification(String callbackUrl, String serviceName, List<ServiceInstance> instances) throws Exception {
        StringBuilder payload = new StringBuilder();
        payload.append("{\"serviceName\":\"").append(serviceName).append("\",");
        payload.append("\"instances\":[");

        for (int i = 0; i < instances.size(); i++) {
            if (i > 0) payload.append(",");
            ServiceInstance inst = instances.get(i);
            payload.append("{");
            payload.append("\"instanceId\":\"").append(inst.getInstanceId()).append("\",");
            payload.append("\"ip\":\"").append(inst.getIp()).append("\",");
            payload.append("\"port\":").append(inst.getPort()).append(",");
            payload.append("\"version\":\"").append(inst.getVersion()).append("\",");
            payload.append("\"zone\":\"").append(inst.getZone()).append("\"");
            payload.append("}");
        }
        payload.append("]}");

        URL url = new URL(callbackUrl);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("POST");
        conn.setRequestProperty("Content-Type", "application/json");
        conn.setDoOutput(true);
        conn.setConnectTimeout(5000);
        conn.setReadTimeout(5000);

        conn.getOutputStream().write(payload.toString().getBytes("UTF-8"));
        conn.getOutputStream().flush();

        int code = conn.getResponseCode();
        conn.disconnect();

        if (code < 200 || code >= 300) {
            throw new Exception("HTTP " + code);
        }
    }

    private String generateInstanceId(ServiceInstance instance) {
        return instance.getServiceName() + "-" + instance.getIp() + "-" + instance.getPort() + "-"
                + UUID.randomUUID().toString().substring(0, 8);
    }
}
