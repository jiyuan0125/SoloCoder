package com.configcenter.service;

import com.configcenter.dto.CallbackNotification;
import com.configcenter.entity.Application;
import com.configcenter.entity.Environment;
import com.configcenter.entity.Watcher;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.web.client.HttpClientErrorException;
import org.springframework.web.client.HttpServerErrorException;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.client.RestTemplate;
import org.springframework.http.*;

import java.util.List;

@Slf4j
@Service
@RequiredArgsConstructor
public class CallbackService {

    private final WatcherService watcherService;
    private final ApplicationService applicationService;
    private final EnvironmentService environmentService;
    private final ObjectMapper objectMapper;

    @Value("${callback.retry.delays:5000,15000}")
    private List<Long> retryDelays;

    @Value("${callback.max-consecutive-failures:5}")
    private int maxConsecutiveFailures;

    @Async
    public void notifyWatchersAsync(Long applicationId, Long environmentId, 
            String configKey, String oldValue, String newValue) {
        try {
            Application app = applicationService.getApplicationById(applicationId);
            Environment env = environmentService.getEnvironmentById(environmentId);
            
            List<Watcher> watchers = watcherService.getWatchersByApplication(applicationId);
            
            if (watchers.isEmpty()) {
                return;
            }

            CallbackNotification notification = new CallbackNotification(
                    app.getName(),
                    env.getName(),
                    configKey,
                    oldValue,
                    newValue
            );

            for (Watcher watcher : watchers) {
                sendNotificationWithRetry(watcher, notification);
            }
        } catch (Exception e) {
            log.error("Error notifying watchers for applicationId={}, key={}", applicationId, configKey, e);
        }
    }

    private void sendNotificationWithRetry(Watcher watcher, CallbackNotification notification) {
        int attempts = 0;
        boolean success = false;
        
        while (attempts <= retryDelays.size()) {
            attempts++;
            try {
                sendNotification(watcher, notification);
                success = true;
                break;
            } catch (Exception e) {
                log.warn("Callback attempt {} failed for watcher {}: {}", 
                        attempts, watcher.getId(), e.getMessage());
                
                if (attempts <= retryDelays.size()) {
                    try {
                        Thread.sleep(retryDelays.get(attempts - 1));
                    } catch (InterruptedException ie) {
                        Thread.currentThread().interrupt();
                        break;
                    }
                }
            }
        }
        
        if (success) {
            watcherService.markWatcherSuccess(watcher);
        } else {
            watcherService.markWatcherFailure(watcher, maxConsecutiveFailures);
            log.error("All callback attempts failed for watcher {}. Consecutive failures: {}/{}", 
                    watcher.getId(), 
                    watcher.getConsecutiveFailures() + 1, 
                    maxConsecutiveFailures);
        }
    }

    private void sendNotification(Watcher watcher, CallbackNotification notification) throws Exception {
        RestTemplate restTemplate = new RestTemplate();
        
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        
        String body = objectMapper.writeValueAsString(notification);
        
        HttpEntity<String> request = new HttpEntity<>(body, headers);
        
        log.debug("Sending callback to {}: {}", watcher.getCallbackUrl(), body);
        
        try {
            ResponseEntity<String> response = restTemplate.exchange(
                    watcher.getCallbackUrl(),
                    HttpMethod.POST,
                    request,
                    String.class
            );
            
            if (!response.getStatusCode().is2xxSuccessful()) {
                throw new RuntimeException("Callback returned non-2xx status: " + response.getStatusCode());
            }
        } catch (HttpClientErrorException | HttpServerErrorException | ResourceAccessException e) {
            throw new RuntimeException("Callback failed: " + e.getMessage(), e);
        }
    }
}
