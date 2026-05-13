package com.messagequeue.dto;

import lombok.Data;

import java.util.Map;

@Data
public class PublishRequest {
    private Map<String, String> headers;
    private String body;
}
