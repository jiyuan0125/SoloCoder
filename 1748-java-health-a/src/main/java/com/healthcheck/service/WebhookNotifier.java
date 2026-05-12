package com.healthcheck.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.healthcheck.model.*;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Instant;
import java.util.HashMap;
import java.util.Map;

@Service
public class WebhookNotifier {

    @Value("${healthcheck.default-webhook:}")
    private String defaultWebhook;

    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;

    public WebhookNotifier(ObjectMapper objectMapper) {
        this.httpClient = HttpClient.newHttpClient();
        this.objectMapper = objectMapper;
    }

    public void notifyServiceStatusChange(String serviceId, String serviceName,
                                          ServiceHealthStatus oldStatus,
                                          ServiceHealthStatus newStatus,
                                          String webhookUrl) {
        Map<String, Object> payload = new HashMap<>();
        payload.put("eventType", "SERVICE_STATUS_CHANGE");
        payload.put("serviceId", serviceId);
        payload.put("serviceName", serviceName);
        payload.put("oldStatus", oldStatus);
        payload.put("newStatus", newStatus);
        payload.put("timestamp", Instant.now().toString());

        sendNotification(payload, webhookUrl);
    }

    public void notifyCheckItemStatusChange(String serviceId, String serviceName,
                                            String checkItemName,
                                            CheckStatus oldStatus,
                                            CheckStatus newStatus,
                                            String webhookUrl) {
        Map<String, Object> payload = new HashMap<>();
        payload.put("eventType", "CHECK_ITEM_STATUS_CHANGE");
        payload.put("serviceId", serviceId);
        payload.put("serviceName", serviceName);
        payload.put("checkItemName", checkItemName);
        payload.put("oldStatus", oldStatus);
        payload.put("newStatus", newStatus);
        payload.put("timestamp", Instant.now().toString());

        sendNotification(payload, webhookUrl);
    }

    private void sendNotification(Map<String, Object> payload, String webhookUrl) {
        String url = (webhookUrl != null && !webhookUrl.isEmpty()) ? webhookUrl : defaultWebhook;

        if (url == null || url.isEmpty()) {
            return;
        }

        try {
            String jsonPayload = objectMapper.writeValueAsString(payload);

            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(url))
                    .header("Content-Type", "application/json")
                    .POST(HttpRequest.BodyPublishers.ofString(jsonPayload))
                    .build();

            httpClient.sendAsync(request, HttpResponse.BodyHandlers.ofString())
                    .exceptionally(e -> {
                        return null;
                    });

        } catch (Exception e) {
        }
    }
}
