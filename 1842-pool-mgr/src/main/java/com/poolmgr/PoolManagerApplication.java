package com.poolmgr;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class PoolManagerApplication {
    public static void main(String[] args) {
        SpringApplication.run(PoolManagerApplication.class, args);
    }
}
