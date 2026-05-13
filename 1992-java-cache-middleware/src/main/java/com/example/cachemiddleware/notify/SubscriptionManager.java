package com.example.cachemiddleware.notify;

import com.example.cachemiddleware.model.KeyChangeEvent;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.scheduling.TaskScheduler;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicInteger;

@Slf4j
@Component
public class SubscriptionManager {

    private final Map<String, Set<Subscriber>> subscribers = new ConcurrentHashMap<>();
    private final RestTemplate restTemplate = new RestTemplate();
    private final ObjectMapper objectMapper = new ObjectMapper();
    private final TaskScheduler taskScheduler;

    private static final int MAX_RETRIES = 2;
    private static final int FAILURES_THRESHOLD = 3;
    private static final long PAUSE_DURATION_MS = 5 * 60 * 1000;

    public SubscriptionManager(TaskScheduler taskScheduler) {
        this.taskScheduler = taskScheduler;
    }

    public void subscribe(String namespace, String callbackUrl) {
        subscribers.computeIfAbsent(namespace, k -> ConcurrentHashMap.newKeySet())
                .add(new Subscriber(callbackUrl));
        log.info("Subscriber registered: namespace={}, url={}", namespace, callbackUrl);
    }

    public void unsubscribe(String namespace, String callbackUrl) {
        Set<Subscriber> subs = subscribers.get(namespace);
        if (subs != null) {
            subs.removeIf(s -> s.callbackUrl.equals(callbackUrl));
            log.info("Subscriber unregistered: namespace={}, url={}", namespace, callbackUrl);
        }
    }

    public void notifyChange(KeyChangeEvent event) {
        Set<Subscriber> subs = subscribers.get(event.getNamespace());
        if (subs == null || subs.isEmpty()) {
            return;
        }
        for (Subscriber subscriber : subs) {
            deliverNotification(subscriber, event);
        }
    }

    private void deliverNotification(Subscriber subscriber, KeyChangeEvent event) {
        if (subscriber.isPaused()) {
            return;
        }

        new Thread(() -> {
            boolean success = false;
            int attempts = 0;
            Exception lastError = null;

            while (attempts <= MAX_RETRIES && !success) {
                try {
                    doCallback(subscriber.callbackUrl, event);
                    success = true;
                    subscriber.consecutiveFailures.set(0);
                } catch (Exception e) {
                    lastError = e;
                    attempts++;
                    if (attempts <= MAX_RETRIES) {
                        try {
                            Thread.sleep(1000);
                        } catch (InterruptedException ie) {
                            Thread.currentThread().interrupt();
                            break;
                        }
                    }
                }
            }

            if (!success) {
                int failures = subscriber.consecutiveFailures.incrementAndGet();
                log.warn("Callback failed: url={}, consecutiveFailures={}, error={}",
                        subscriber.callbackUrl, failures, lastError != null ? lastError.getMessage() : "unknown");
                if (failures >= FAILURES_THRESHOLD) {
                    pauseSubscriber(subscriber);
                }
            }
        }).start();
    }

    private void doCallback(String url, KeyChangeEvent event) throws Exception {
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        String body = objectMapper.writeValueAsString(event);
        HttpEntity<String> entity = new HttpEntity<>(body, headers);
        restTemplate.postForEntity(url, entity, String.class);
    }

    private void pauseSubscriber(Subscriber subscriber) {
        subscriber.pausedUntil = System.currentTimeMillis() + PAUSE_DURATION_MS;
        log.warn("Subscriber paused for 5 minutes: url={}", subscriber.callbackUrl);
        taskScheduler.schedule(() -> {
            subscriber.pausedUntil = 0;
            subscriber.consecutiveFailures.set(0);
            log.info("Subscriber resumed: url={}", subscriber.callbackUrl);
        }, new Date(subscriber.pausedUntil));
    }

    public static class Subscriber {
        final String callbackUrl;
        final AtomicInteger consecutiveFailures = new AtomicInteger(0);
        volatile long pausedUntil = 0;

        public Subscriber(String callbackUrl) {
            this.callbackUrl = callbackUrl;
        }

        public boolean isPaused() {
            return pausedUntil > System.currentTimeMillis();
        }

        @Override
        public boolean equals(Object o) {
            if (this == o) return true;
            if (o == null || getClass() != o.getClass()) return false;
            Subscriber that = (Subscriber) o;
            return Objects.equals(callbackUrl, that.callbackUrl);
        }

        @Override
        public int hashCode() {
            return Objects.hash(callbackUrl);
        }
    }
}
