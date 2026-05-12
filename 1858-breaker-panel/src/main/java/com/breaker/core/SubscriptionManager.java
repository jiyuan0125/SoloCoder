package com.breaker.core;

import com.breaker.model.NotificationPayload;
import com.breaker.model.StateChangeEvent;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import javax.annotation.PostConstruct;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

@Component
public class SubscriptionManager {
    private static final Logger logger = LoggerFactory.getLogger(SubscriptionManager.class);
    private static final int MAX_RETRIES = 2;

    private final Map<String, List<String>> subscriptions = new ConcurrentHashMap<>();
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;
    private final ExecutorService executor;

    public SubscriptionManager(ObjectMapper objectMapper) {
        this.httpClient = HttpClient.newHttpClient();
        this.objectMapper = objectMapper;
        this.executor = Executors.newCachedThreadPool();
    }

    @PostConstruct
    public void init() {
    }

    public void addSubscription(String serviceName, String callbackUrl) {
        subscriptions.compute(serviceName, (name, existing) -> {
            List<String> urls = existing != null ? existing : new ArrayList<>();
            if (!urls.contains(callbackUrl)) {
                urls.add(callbackUrl);
            }
            return urls;
        });
    }

    public void notifyStateChange(String serviceName, StateChangeEvent event) {
        List<String> urls = subscriptions.get(serviceName);
        if (urls == null || urls.isEmpty()) {
            return;
        }

        NotificationPayload payload = new NotificationPayload(serviceName, event);
        for (String url : urls) {
            executor.submit(() -> sendNotificationWithRetry(url, payload));
        }
    }

    private void sendNotificationWithRetry(String url, NotificationPayload payload) {
        int attempt = 0;
        while (attempt <= MAX_RETRIES) {
            try {
                String jsonBody = objectMapper.writeValueAsString(payload);
                HttpRequest request = HttpRequest.newBuilder()
                        .uri(URI.create(url))
                        .header("Content-Type", "application/json")
                        .POST(HttpRequest.BodyPublishers.ofString(jsonBody))
                        .build();

                HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
                int status = response.statusCode();

                if (status >= 200 && status < 300) {
                    logger.info("Notification sent successfully to {} for service {}", url, payload.getService());
                    return;
                }

                logger.warn("Notification failed with status {} for service {}, attempt {}/{}",
                        status, payload.getService(), attempt + 1, MAX_RETRIES + 1);

            } catch (Exception e) {
                logger.warn("Notification exception for service {}, attempt {}/{}: {}",
                        payload.getService(), attempt + 1, MAX_RETRIES + 1, e.getMessage());
            }

            attempt++;
            if (attempt <= MAX_RETRIES) {
                try {
                    Thread.sleep(1000);
                } catch (InterruptedException ie) {
                    Thread.currentThread().interrupt();
                    return;
                }
            }
        }

        logger.error("Notification failed after {} retries for service {} to url {}",
                MAX_RETRIES + 1, payload.getService(), url);
    }
}
