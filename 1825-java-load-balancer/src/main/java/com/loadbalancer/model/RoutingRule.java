package com.loadbalancer.model;

import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.util.ArrayList;
import java.util.List;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class RoutingRule {
    private String id;
    private String pathPattern;
    @Builder.Default
    private List<String> requiredTags = new ArrayList<>();
    private String description;
    private int priority;
}
