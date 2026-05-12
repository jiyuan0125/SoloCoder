package com.configcenter.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix = "config.long-poll")
@Data
public class LongPollConfig {
    
    private int timeoutSeconds = 30;
}
