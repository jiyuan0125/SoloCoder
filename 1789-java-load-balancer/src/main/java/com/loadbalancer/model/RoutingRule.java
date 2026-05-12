package com.loadbalancer.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

import java.util.Set;
import java.util.HashSet;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class RoutingRule {
    private String id;
    private String pathPattern;
    private Set<String> requiredTags = new HashSet<>();
    private int priority = 0;
    private boolean enabled = true;

    public RoutingRule(String id, String pathPattern, Set<String> requiredTags) {
        this.id = id;
        this.pathPattern = pathPattern;
        this.requiredTags = requiredTags != null ? requiredTags : new HashSet<>();
    }
}
