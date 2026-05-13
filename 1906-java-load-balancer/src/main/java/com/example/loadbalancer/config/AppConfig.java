package com.example.loadbalancer.config;

import com.example.loadbalancer.strategy.StrategyManager;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.web.client.RestTemplateBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.client.RestTemplate;

import java.time.Duration;

@Configuration
public class AppConfig {

    @Value("${load-balancer.health-check.timeout-ms:3000}")
    private int healthCheckTimeoutMs;

    @Value("${load-balancer.initial-strategy:round_robin}")
    private String initialStrategy;

    @Bean
    public RestTemplate restTemplate(RestTemplateBuilder builder) {
        return builder
                .setConnectTimeout(Duration.ofMillis(healthCheckTimeoutMs))
                .setReadTimeout(Duration.ofMillis(healthCheckTimeoutMs))
                .build();
    }

    @Bean
    public CommandLineRunner initializeStrategy(StrategyManager strategyManager) {
        return args -> {
            strategyManager.setInitialStrategy(initialStrategy);
        };
    }
}
