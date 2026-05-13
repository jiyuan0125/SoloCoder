package com.example.apiproxymirror.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

@Data
@Configuration
@ConfigurationProperties(prefix = "proxy")
public class ProxyConfig {
    private String targetUrl;
    private int timeout = 10000;
}
