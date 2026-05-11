package com.retail.memberpoints;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class MemberPointsApplication {

    public static void main(String[] args) {
        SpringApplication.run(MemberPointsApplication.class, args);
    }
}
