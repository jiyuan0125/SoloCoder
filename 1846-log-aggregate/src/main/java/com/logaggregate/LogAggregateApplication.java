package com.logaggregate;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class LogAggregateApplication {

    public static void main(String[] args) {
        SpringApplication.run(LogAggregateApplication.class, args);
    }
}
