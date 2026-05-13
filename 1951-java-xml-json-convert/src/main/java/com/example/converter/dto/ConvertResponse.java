package com.example.converter.dto;

import lombok.Data;

@Data
public class ConvertResponse {
    private String result;
    private String format;
    private Boolean validated;
}
