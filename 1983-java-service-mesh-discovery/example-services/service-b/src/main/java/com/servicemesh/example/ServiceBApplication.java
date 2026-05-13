package com.servicemesh.example;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@SpringBootApplication
@RestController
public class ServiceBApplication {

    public static void main(String[] args) {
        SpringApplication.run(ServiceBApplication.class, args);
    }

    @GetMapping("/hello")
    public String hello() {
        return "Hello from Service B (v2)";
    }

    @GetMapping("/api/process")
    public String process() {
        return "Service B processed your request";
    }
}
