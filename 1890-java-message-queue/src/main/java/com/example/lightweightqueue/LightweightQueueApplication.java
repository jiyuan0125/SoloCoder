package com.example.lightweightqueue;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class LightweightQueueApplication {

    public static void main(String[] args) {
        SpringApplication.run(LightweightQueueApplication.class, args);
    }
}
