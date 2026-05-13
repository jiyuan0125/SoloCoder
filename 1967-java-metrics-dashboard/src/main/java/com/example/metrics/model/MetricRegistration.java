package com.example.metrics.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

import javax.validation.constraints.NotBlank;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class MetricRegistration {
    @NotBlank(message = "name is required")
    private String name;

    private String description;
}
