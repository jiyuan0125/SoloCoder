package com.example.gateway;

import com.example.gateway.model.BackendTarget;
import com.example.gateway.model.RouteRule;
import com.example.gateway.service.RouteService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.CommandLineRunner;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

import java.util.Arrays;

@SpringBootApplication
public class GatewayApplication implements CommandLineRunner {

    @Autowired
    private RouteService routeService;

    public static void main(String[] args) {
        SpringApplication.run(GatewayApplication.class, args);
    }

    @Override
    public void run(String... args) throws Exception {
        RouteRule publicRoute = new RouteRule(
            "public-api",
            "/api/public/",
            Arrays.asList(new BackendTarget("http://localhost:8081", 1)),
            Arrays.asList(),
            10
        );
        routeService.addRoute(publicRoute);

        RouteRule adminRoute = new RouteRule(
            "admin-api",
            "/api/admin/",
            Arrays.asList(new BackendTarget("http://localhost:8082", 1)),
            Arrays.asList("auth", "permission", "rateLimit"),
            20
        );
        routeService.addRoute(adminRoute);

        RouteRule internalRoute = new RouteRule(
            "internal-api",
            "/api/internal/",
            Arrays.asList(new BackendTarget("http://localhost:8083", 1)),
            Arrays.asList("auth", "permission"),
            15
        );
        routeService.addRoute(internalRoute);
    }
}
