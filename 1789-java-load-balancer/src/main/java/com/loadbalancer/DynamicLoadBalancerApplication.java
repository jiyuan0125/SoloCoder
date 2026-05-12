package com.loadbalancer;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class DynamicLoadBalancerApplication {
    public static void main(String[] args) {
        SpringApplication.run(DynamicLoadBalancerApplication.class, args);
    }
}
