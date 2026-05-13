package com.example.apiproxymirror.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Data
@Configuration
@ConfigurationProperties(prefix = "recording")
public class RecordingConfig {
    private int maxRecords = 5000;
    private boolean enabled = false;
}
