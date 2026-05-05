package com.training.server;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class TrainingServerApplication {

    public static void main(String[] args) {
        SpringApplication.run(TrainingServerApplication.class, args);
    }
}
