package com.loadbalancer.dto;

import lombok.Data;

import java.util.ArrayList;
import java.util.List;

@Data
public class RoutingRuleRequest {
    private String pathPattern;
    private List<String> requiredTags = new ArrayList<>();
    private String description;
    private int priority = 0;
}
