package com.example.pool;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class ConnectionPoolManagerApplication {

    public static void main(String[] args) {
        SpringApplication.run(ConnectionPoolManagerApplication.class, args);
    }
}
