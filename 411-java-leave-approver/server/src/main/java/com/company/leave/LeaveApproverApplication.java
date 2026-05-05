package com.company.leave;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class LeaveApproverApplication {
    public static void main(String[] args) {
        SpringApplication.run(LeaveApproverApplication.class, args);
    }
}
