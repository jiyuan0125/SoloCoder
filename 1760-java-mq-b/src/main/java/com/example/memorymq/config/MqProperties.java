package com.example.memorymq.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "mq")
public class MqProperties {
    private int defaultMaxMessagesPerTopic = 100000;
    private int defaultMaxRetry = 3;
    private int retryIntervalSeconds = 5;
    private int backlogCheckIntervalSeconds = 10;
    private int ackTimeoutSeconds = 60;
    private int retentionMinutes = 1440;
}
