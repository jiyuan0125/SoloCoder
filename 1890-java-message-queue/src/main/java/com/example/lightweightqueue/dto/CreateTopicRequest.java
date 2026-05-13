package com.example.lightweightqueue.dto;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class CreateTopicRequest {
    @NotBlank
    private String name;
    
    @Min(1)
    private int maxCapacity;
}
