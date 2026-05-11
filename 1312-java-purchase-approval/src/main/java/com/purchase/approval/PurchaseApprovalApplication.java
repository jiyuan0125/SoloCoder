package com.purchase.approval;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class PurchaseApprovalApplication {
    public static void main(String[] args) {
        SpringApplication.run(PurchaseApprovalApplication.class, args);
    }
}
