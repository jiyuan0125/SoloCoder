package com.example.messagequeue;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class MessageQueueApplication {

    public static void main(String[] args) {
        SpringApplication.run(MessageQueueApplication.class, args);
    }
}
