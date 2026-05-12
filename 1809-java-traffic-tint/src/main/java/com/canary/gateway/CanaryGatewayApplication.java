package com.canary.gateway;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class CanaryGatewayApplication {
    public static void main(String[] args) {
        SpringApplication.run(CanaryGatewayApplication.class, args);
    }
}
