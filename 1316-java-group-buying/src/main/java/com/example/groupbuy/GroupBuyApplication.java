package com.example.groupbuy;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class GroupBuyApplication {
    public static void main(String[] args) {
        SpringApplication.run(GroupBuyApplication.class, args);
    }
}
