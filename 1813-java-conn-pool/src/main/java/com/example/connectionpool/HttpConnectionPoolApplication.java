package com.example.connectionpool;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class HttpConnectionPoolApplication {

    public static void main(String[] args) {
        SpringApplication.run(HttpConnectionPoolApplication.class, args);
    }
}
