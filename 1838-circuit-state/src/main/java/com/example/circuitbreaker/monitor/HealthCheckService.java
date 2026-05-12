package com.example.circuitbreaker.monitor;

import com.example.circuitbreaker.breaker.CircuitBreaker;
import com.example.circuitbreaker.breaker.CircuitBreakerRegistry;
import com.example.circuitbreaker.breaker.CircuitBreakerState;
import com.example.circuitbreaker.config.ConfigService;
import com.example.circuitbreaker.stats.ChangeReason;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.Map;

@Slf4j
@Service
@RequiredArgsConstructor
public class HealthCheckService {

    private final CircuitBreakerRegistry circuitBreakerRegistry;
    private final ConfigService configService;
    private final RestTemplate restTemplate;

    @Value("${monitor.interval:10000}")
    private long monitorInterval;

    @Scheduled(fixedRateString = "${monitor.interval:10000}")
    public void checkAllServices() {
        Map<String, CircuitBreaker> allBreakers = circuitBreakerRegistry.getAll();
        
        for (Map.Entry<String, CircuitBreaker> entry : allBreakers.entrySet()) {
            String serviceName = entry.getKey();
            CircuitBreaker breaker = entry.getValue();
            
            try {
                checkService(serviceName, breaker);
            } catch (Exception e) {
                log.error("Error checking service {}: {}", serviceName, e.getMessage());
            }
        }
    }

    private void checkService(String serviceName, CircuitBreaker breaker) {
        String healthUrl = "http://" + serviceName + "/health";
        
        try {
            ResponseEntity<String> response = restTemplate.getForEntity(healthUrl, String.class);
            if (response.getStatusCode().is2xxSuccessful()) {
                log.debug("Service {} health check: OK", serviceName);
                if (breaker.getState() == CircuitBreakerState.OPEN) {
                    breaker.transitionTo(CircuitBreakerState.HALF_OPEN, ChangeReason.HEALTH_CHECK);
                }
            } else {
                log.debug("Service {} health check: FAILED (status: {})", serviceName, response.getStatusCode());
            }
        } catch (Exception e) {
            log.debug("Service {} health check: FAILED (error: {})", serviceName, e.getMessage());
        }
    }
}
