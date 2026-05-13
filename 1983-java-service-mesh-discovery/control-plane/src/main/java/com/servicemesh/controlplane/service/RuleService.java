package com.servicemesh.controlplane.service;

import com.servicemesh.common.model.TrafficRule;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import javax.annotation.PostConstruct;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class RuleService {

    private static final Logger log = LoggerFactory.getLogger(RuleService.class);

    private final Map<String, TrafficRule> rules = new ConcurrentHashMap<>();

    public void setRule(TrafficRule rule) {
        rule.setLastUpdate(System.currentTimeMillis());
        rules.put(rule.getServiceName(), rule);
        log.info("Set traffic rule for {}: type={}", rule.getServiceName(), rule.getType());
    }

    public void removeRule(String serviceName) {
        rules.remove(serviceName);
        log.info("Removed traffic rule for {}", serviceName);
    }

    public TrafficRule getRule(String serviceName) {
        return rules.get(serviceName);
    }

    public List<TrafficRule> getAllRules() {
        return new ArrayList<>(rules.values());
    }

    @PostConstruct
    public void init() {
        log.info("RuleService initialized");
    }
}
