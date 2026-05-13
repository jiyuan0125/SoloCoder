package com.example.converter.dto;

import lombok.Data;

@Data
public class ConvertRequest {
    private String data;
    private String mappingName;
    private Boolean validate = false;
}
