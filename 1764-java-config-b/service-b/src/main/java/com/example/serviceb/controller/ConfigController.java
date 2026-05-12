package com.example.serviceb.controller;

import com.example.serviceb.config.ServiceBConfig;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.cloud.context.config.annotation.RefreshScope;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/api")
@RefreshScope
public class ConfigController {

    @Autowired
    private ServiceBConfig serviceBConfig;

    @Value("${app.common.environment:default}")
    private String environment;

    @Value("${app.common.version:unknown}")
    private String version;

    @Value("${feature.service-b.enable-notifications:false}")
    private boolean notificationsEnabled;

    @Value("${feature.service-b.notification-template:default}")
    private String notificationTemplate;

    @GetMapping("/config")
    public Map<String, Object> getConfig() {
        Map<String, Object> config = new HashMap<>();
        config.put("environment", environment);
        config.put("version", version);
        config.put("notificationsEnabled", notificationsEnabled);
        config.put("notificationTemplate", notificationTemplate);
        config.put("serviceConfig", serviceBConfig);
        return config;
    }

    @GetMapping("/health")
    public Map<String, String> health() {
        Map<String, String> health = new HashMap<>();
        health.put("status", "UP");
        health.put("service", "Service B");
        health.put("environment", environment);
        return health;
    }
}
