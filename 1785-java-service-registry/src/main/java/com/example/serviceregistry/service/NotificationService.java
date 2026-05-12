package com.example.serviceregistry.service;

import com.example.serviceregistry.event.ServiceChangedEvent;
import com.example.serviceregistry.model.ServiceInstance;
import com.example.serviceregistry.model.Subscriber;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestClient;

import java.util.List;
import java.util.Set;

@Slf4j
@Service
@RequiredArgsConstructor
public class NotificationService {

    private final ServiceRegistry serviceRegistry;
    private final RestClient restClient;

    @EventListener
    public void handleServiceChanged(ServiceChangedEvent event) {
        String serviceName = event.getServiceName();
        List<ServiceInstance> healthyInstances = event.getHealthyInstances();

        Set<Subscriber> subscribers = serviceRegistry.getSubscribers(serviceName);
        if (subscribers.isEmpty()) {
            return;
        }

        log.info("Notifying {} subscribers for service: {}, {} healthy instances",
                subscribers.size(), serviceName, healthyInstances.size());

        subscribers.forEach(subscriber -> notifySubscriber(subscriber, healthyInstances));
    }

    private void notifySubscriber(Subscriber subscriber, List<ServiceInstance> instances) {
        try {
            restClient.post()
                    .uri(subscriber.getCallbackUrl())
                    .body(instances)
                    .retrieve()
                    .toBodilessEntity();
            log.debug("Successfully notified subscriber: {}", subscriber.getSubscriberId());
        } catch (Exception e) {
            log.warn("Failed to notify subscriber {} at {}: {}",
                    subscriber.getSubscriberId(), subscriber.getCallbackUrl(), e.getMessage());
        }
    }
}
