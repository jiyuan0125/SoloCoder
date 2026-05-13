package com.servicemesh.example;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;

@SpringBootApplication
@RestController
public class ServiceAApplication {

    public static void main(String[] args) {
        SpringApplication.run(ServiceAApplication.class, args);
    }

    @GetMapping("/hello")
    public String hello() {
        return "Hello from Service A (v1)";
    }

    @GetMapping("/api/data")
    public String getData() {
        return "Service A data response";
    }
}
