package com.example.loadbalancer.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;

@Data
public class StrategyRequest {
    @NotBlank(message = "Strategy is required")
    private String strategy;
}
