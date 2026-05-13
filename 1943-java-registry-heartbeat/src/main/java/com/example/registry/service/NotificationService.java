package com.example.registry.service;

import com.example.registry.config.RegistryProperties;
import com.example.registry.model.ChangeEvent;
import com.example.registry.model.Subscription;
import com.example.registry.store.SubscriptionStore;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpStatus;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.time.Duration;
import java.util.List;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

@Slf4j
@Service
@RequiredArgsConstructor
public class NotificationService {
    private final SubscriptionStore subscriptionStore;
    private final RegistryProperties registryProperties;
    private final ObjectMapper objectMapper = new ObjectMapper();
    private final ExecutorService executorService = Executors.newCachedThreadPool();

    public void notifySubscribers(String serviceName, ChangeEvent event) {
        List<Subscription> subscriptions = subscriptionStore.findByServiceName(serviceName);
        if (subscriptions.isEmpty()) {
            return;
        }

        for (Subscription subscription : subscriptions) {
            executorService.submit(() -> sendNotificationWithRetry(subscription, event));
        }
    }

    private void sendNotificationWithRetry(Subscription subscription, ChangeEvent event) {
        int maxRetries = registryProperties.getNotificationRetryMax();
        long retryInterval = registryProperties.getNotificationRetryIntervalMs();
        long timeout = registryProperties.getNotificationTimeoutMs();

        RestTemplate restTemplate = createRestTemplate(timeout);

        for (int attempt = 0; attempt <= maxRetries; attempt++) {
            try {
                String body = objectMapper.writeValueAsString(event);
                log.debug("Sending notification to {} (attempt {}/{}): {}", 
                    subscription.getCallbackUrl(), attempt + 1, maxRetries + 1, body);

                org.springframework.http.HttpEntity<String> requestEntity = 
                    new org.springframework.http.HttpEntity<>(body, createHeaders());

                org.springframework.http.ResponseEntity<String> response = restTemplate.postForEntity(
                    subscription.getCallbackUrl(), requestEntity, String.class);

                if (HttpStatus.valueOf(response.getStatusCode().value()).is2xxSuccessful()) {
                    log.debug("Notification sent successfully to {}", subscription.getCallbackUrl());
                    return;
                } else {
                    log.warn("Notification failed for {} with status {}", 
                        subscription.getCallbackUrl(), response.getStatusCode());
                }
            } catch (Exception e) {
                log.warn("Notification failed for {}: {}", subscription.getCallbackUrl(), e.getMessage());
            }

            if (attempt < maxRetries) {
                try {
                    Thread.sleep(retryInterval);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    return;
                }
            }
        }

        log.error("Notification failed after {} attempts for {}", 
            maxRetries + 1, subscription.getCallbackUrl());
    }

    private RestTemplate createRestTemplate(long timeoutMs) {
        SimpleClientHttpRequestFactory factory = new SimpleClientHttpRequestFactory();
        factory.setConnectTimeout(Duration.ofMillis(timeoutMs));
        factory.setReadTimeout(Duration.ofMillis(timeoutMs));
        return new RestTemplate(factory);
    }

    private org.springframework.http.HttpHeaders createHeaders() {
        org.springframework.http.HttpHeaders headers = new org.springframework.http.HttpHeaders();
        headers.setContentType(org.springframework.http.MediaType.APPLICATION_JSON);
        return headers;
    }
}
