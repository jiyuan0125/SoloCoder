package com.example.healthchecker.model;

import lombok.Data;

@Data
public class AddDependencyRequest {
    private String dependency_service_name;
}
