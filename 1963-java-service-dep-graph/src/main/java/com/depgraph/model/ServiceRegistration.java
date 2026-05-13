package com.depgraph.model;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.NotNull;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ServiceRegistration {
    @NotBlank(message = "Service name is required")
    private String serviceName;

    @NotNull(message = "Dependencies list is required")
    private List<String> dependencies;
}
