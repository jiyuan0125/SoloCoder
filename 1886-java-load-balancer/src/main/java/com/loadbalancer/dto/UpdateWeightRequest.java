package com.loadbalancer.dto;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class UpdateWeightRequest {
    @NotNull(message = "Weight is required")
    @Min(value = 0, message = "Weight must be >= 0")
    private Integer weight;
}
