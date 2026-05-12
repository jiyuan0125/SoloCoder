package com.example.memorymq.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class CreateTopicRequest {
    @NotBlank
    private String name;
    private Integer maxMessages;
}
