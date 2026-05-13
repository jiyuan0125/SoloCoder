package com.servicemesh.sidecar.service;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.servicemesh.sidecar.config.SidecarProperties;
import okhttp3.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import javax.annotation.PreDestroy;
import java.io.IOException;
import java.util.Map;
import java.util.concurrent.TimeUnit;

@Service
public class RegistrationService {

    private static final Logger log = LoggerFactory.getLogger(RegistrationService.class);
    private static final MediaType JSON = MediaType.parse("application/json; charset=utf-8");

    private final SidecarProperties properties;
    private final OkHttpClient httpClient;
    private final ObjectMapper objectMapper;

    private String instanceId;
    private volatile boolean registered = false;

    public RegistrationService(SidecarProperties properties) {
        this.properties = properties;
        this.httpClient = new OkHttpClient.Builder()
                .connectTimeout(5, TimeUnit.SECONDS)
                .readTimeout(5, TimeUnit.SECONDS)
                .build();
        this.objectMapper = new ObjectMapper();
    }

    @PostConstruct
    public void init() {
        log.info("Sidecar starting for service: {} ({}:{}, version={}, zone={})",
                properties.getServiceName(), properties.getAppHost(), properties.getAppPort(),
                properties.getVersion(), properties.getZone());
        register();
    }

    public void register() {
        try {
            Map<String, Object> payload = Map.of(
                    "serviceName", properties.getServiceName(),
                    "ip", properties.getLocalIp(),
                    "port", properties.getAppPort(),
                    "version", properties.getVersion(),
                    "zone", properties.getZone()
            );

            RequestBody body = RequestBody.create(objectMapper.writeValueAsString(payload), JSON);
            Request request = new Request.Builder()
                    .url(properties.getRegistryUrl() + "/api/registry/register")
                    .post(body)
                    .build();

            try (Response response = httpClient.newCall(request).execute()) {
                if (response.isSuccessful() && response.body() != null) {
                    JsonNode json = objectMapper.readTree(response.body().string());
                    instanceId = json.get("instanceId").asText();
                    registered = true;
                    log.info("Service registered successfully, instanceId: {}", instanceId);
                } else {
                    log.error("Registration failed: HTTP {}", response.code());
                }
            }
        } catch (Exception e) {
            log.error("Registration error", e);
        }
    }

    @Scheduled(fixedRateString = "${sidecar.heartbeat-interval:5000}")
    public void sendHeartbeat() {
        if (!registered || instanceId == null) {
            register();
            return;
        }

        try {
            Map<String, String> payload = Map.of(
                    "serviceName", properties.getServiceName(),
                    "instanceId", instanceId
            );

            RequestBody body = RequestBody.create(objectMapper.writeValueAsString(payload), JSON);
            Request request = new Request.Builder()
                    .url(properties.getRegistryUrl() + "/api/registry/heartbeat")
                    .post(body)
                    .build();

            try (Response response = httpClient.newCall(request).execute()) {
                if (response.code() == 404) {
                    log.warn("Instance not found, re-registering...");
                    registered = false;
                    register();
                }
            }
        } catch (IOException e) {
            log.debug("Heartbeat failed: {}", e.getMessage());
        }
    }

    @PreDestroy
    public void deregister() {
        if (!registered || instanceId == null) return;

        try {
            Map<String, String> payload = Map.of(
                    "serviceName", properties.getServiceName(),
                    "instanceId", instanceId
            );

            RequestBody body = RequestBody.create(objectMapper.writeValueAsString(payload), JSON);
            Request request = new Request.Builder()
                    .url(properties.getRegistryUrl() + "/api/registry/deregister")
                    .post(body)
                    .build();

            try (Response response = httpClient.newCall(request).execute()) {
                log.info("Service deregistered: {}", response.isSuccessful());
            }
        } catch (IOException e) {
            log.warn("Deregistration failed", e);
        }
    }

    public boolean isRegistered() { return registered; }
    public String getInstanceId() { return instanceId; }
}
