package com.configcenter.client;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;
import java.util.function.Consumer;

@Slf4j
public class ConfigCenterClient {

    private final String baseUrl;
    private final Long projectId;
    private final String environment;
    private final String instanceId;
    private final ObjectMapper objectMapper;
    
    private String currentVersion;
    private Map<String, String> currentConfig = new HashMap<>();
    private final ScheduledExecutorService executorService;
    private Consumer<Map<String, String>> configChangeListener;

    public ConfigCenterClient(String baseUrl, Long projectId, String environment) {
        this.baseUrl = baseUrl;
        this.projectId = projectId;
        this.environment = environment;
        this.instanceId = UUID.randomUUID().toString();
        this.objectMapper = new ObjectMapper();
        this.executorService = Executors.newSingleThreadScheduledExecutor();
    }

    public void start() {
        try {
            fetchInitialConfig();
            startLongPolling();
            log.info("ConfigCenterClient started for project {} environment {}", projectId, environment);
        } catch (Exception e) {
            log.error("Failed to start ConfigCenterClient", e);
            throw new RuntimeException("Failed to start ConfigCenterClient", e);
        }
    }

    public void stop() {
        executorService.shutdown();
        try {
            if (!executorService.awaitTermination(5, TimeUnit.SECONDS)) {
                executorService.shutdownNow();
            }
        } catch (InterruptedException e) {
            executorService.shutdownNow();
            Thread.currentThread().interrupt();
        }
        log.info("ConfigCenterClient stopped");
    }

    public void setConfigChangeListener(Consumer<Map<String, String>> listener) {
        this.configChangeListener = listener;
    }

    public Map<String, String> getCurrentConfig() {
        return new HashMap<>(currentConfig);
    }

    public String getConfig(String key) {
        return currentConfig.get(key);
    }

    private void fetchInitialConfig() throws Exception {
        String url = baseUrl + "/api/projects/" + projectId + "/configs/" + environment + "/released";
        String response = sendGetRequest(url);
        
        JsonNode root = objectMapper.readTree(response);
        JsonNode dataNode = root.get("data");
        
        Map<String, String> configMap = new HashMap<>();
        if (dataNode.isArray()) {
            for (JsonNode configNode : dataNode) {
                String key = configNode.get("configKey").asText();
                String value = configNode.get("currentValue").asText();
                configMap.put(key, value);
            }
        }
        
        this.currentConfig = configMap;
        this.currentVersion = calculateVersion(configMap);
        
        log.info("Initial config fetched: {} items, version: {}", currentConfig.size(), currentVersion);
    }

    private void startLongPolling() {
        executorService.scheduleWithFixedDelay(() -> {
            try {
                String url = baseUrl + "/api/projects/" + projectId + "/subscriptions/watch";
                
                Map<String, Object> requestBody = new HashMap<>();
                requestBody.put("instanceId", instanceId);
                requestBody.put("environment", environment);
                requestBody.put("lastKnownVersion", currentVersion);
                
                String response = sendPostRequest(url, requestBody);
                
                JsonNode root = objectMapper.readTree(response);
                boolean success = root.get("success").asBoolean();
                JsonNode dataNode = root.get("data");
                
                if (success && dataNode != null && !dataNode.isNull()) {
                    JsonNode changesNode = dataNode.get("changes");
                    Map<String, String> newConfig = new HashMap<>();
                    
                    if (changesNode != null && changesNode.isObject()) {
                        changesNode.fields().forEachRemaining(entry -> {
                            newConfig.put(entry.getKey(), entry.getValue().asText());
                        });
                    }
                    
                    this.currentConfig = newConfig;
                    this.currentVersion = calculateVersion(newConfig);
                    
                    log.info("Config updated. New version: {}, Config size: {}", currentVersion, currentConfig.size());
                    
                    if (configChangeListener != null) {
                        configChangeListener.accept(newConfig);
                    }
                }
            } catch (Exception e) {
                log.error("Error during long polling", e);
            }
        }, 0, 1, TimeUnit.MILLISECONDS);
    }

    private String sendGetRequest(String urlString) throws Exception {
        URL url = new URL(urlString);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("GET");
        conn.setConnectTimeout(5000);
        conn.setReadTimeout(60000);
        
        return readResponse(conn);
    }

    private String sendPostRequest(String urlString, Map<String, Object> body) throws Exception {
        URL url = new URL(urlString);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("POST");
        conn.setRequestProperty("Content-Type", "application/json");
        conn.setConnectTimeout(5000);
        conn.setReadTimeout(60000);
        conn.setDoOutput(true);
        
        String jsonBody = objectMapper.writeValueAsString(body);
        try (OutputStream os = conn.getOutputStream()) {
            os.write(jsonBody.getBytes(StandardCharsets.UTF_8));
            os.flush();
        }
        
        return readResponse(conn);
    }

    private String readResponse(HttpURLConnection conn) throws Exception {
        int responseCode = conn.getResponseCode();
        BufferedReader reader;
        
        if (responseCode >= 200 && responseCode < 300) {
            reader = new BufferedReader(new InputStreamReader(conn.getInputStream(), StandardCharsets.UTF_8));
        } else {
            reader = new BufferedReader(new InputStreamReader(conn.getErrorStream(), StandardCharsets.UTF_8));
        }
        
        StringBuilder response = new StringBuilder();
        String line;
        while ((line = reader.readLine()) != null) {
            response.append(line);
        }
        reader.close();
        
        return response.toString();
    }

    private String calculateVersion(Map<String, String> config) {
        try {
            return Integer.toHexString(objectMapper.writeValueAsString(config).hashCode());
        } catch (Exception e) {
            return String.valueOf(System.currentTimeMillis());
        }
    }
}
