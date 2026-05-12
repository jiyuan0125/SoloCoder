package com.trace.server.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "trace")
public class TraceConfig {
    private double sampleRate = 1.0;
    private int serverPort = 8080;
}
