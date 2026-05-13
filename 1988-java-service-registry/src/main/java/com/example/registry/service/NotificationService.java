package com.example.registry.service;

import com.example.registry.config.RegistryProperties;
import com.example.registry.model.ServiceInstance;
import com.example.registry.model.Subscriber;
import com.example.registry.repository.ServiceRegistryRepository;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.List;

@Slf4j
@Service
public class NotificationService {

    private final ServiceRegistryRepository repository;
    private final RegistryProperties properties;
    private final RestTemplate restTemplate;
    private final ObjectMapper objectMapper;

    public NotificationService(ServiceRegistryRepository repository,
                               RegistryProperties properties,
                               RestTemplate restTemplate,
                               ObjectMapper objectMapper) {
        this.repository = repository;
        this.properties = properties;
        this.restTemplate = restTemplate;
        this.objectMapper = objectMapper;
    }

    @Async
    public void notifyMetadataChange(String serviceName, ServiceInstance instance) {
        List<Subscriber> serviceSubscribers = repository.getSubscribers(serviceName);
        if (serviceSubscribers.isEmpty()) {
            return;
        }

        for (Subscriber subscriber : serviceSubscribers) {
            try {
                sendNotification(subscriber, instance);
                subscriber.setConsecutiveFailures(0);
            } catch (Exception e) {
                log.warn("Failed to notify subscriber {} for service {}: {}",
                    subscriber.getCallbackUrl(), serviceName, e.getMessage());
                subscriber.setConsecutiveFailures(subscriber.getConsecutiveFailures() + 1);
                
                if (subscriber.getConsecutiveFailures() >= properties.getNotification().getMaxRetries()) {
                    log.info("Removing subscriber {} due to {} consecutive failures",
                        subscriber.getCallbackUrl(), subscriber.getConsecutiveFailures());
                    repository.unsubscribe(serviceName, subscriber.getCallbackUrl());
                }
            }
        }
    }

    private void sendNotification(Subscriber subscriber, ServiceInstance instance) throws Exception {
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        
        String body = objectMapper.writeValueAsString(instance);
        HttpEntity<String> request = new HttpEntity<>(body, headers);
        
        restTemplate.postForEntity(subscriber.getCallbackUrl(), request, String.class);
    }
}
