package com.example.cachemiddleware.model;

import lombok.Data;

import javax.validation.constraints.NotEmpty;
import java.util.Map;

@Data
public class BatchWriteRequest {
    @NotEmpty(message = "entries cannot be empty")
    private Map<String, Object> entries;
}
