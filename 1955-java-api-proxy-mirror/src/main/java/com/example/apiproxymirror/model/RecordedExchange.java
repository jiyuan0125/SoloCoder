package com.example.apiproxymirror.model;

import lombok.Builder;
import lombok.Data;

import java.time.Instant;
import java.util.Map;

@Data
@Builder
public class RecordedExchange {
    private String id;
    private String method;
    private String path;
    private Map<String, String> requestHeaders;
    private String requestBody;
    private int responseStatus;
    private Map<String, String> responseHeaders;
    private String responseBody;
    private Instant timestamp;
}
