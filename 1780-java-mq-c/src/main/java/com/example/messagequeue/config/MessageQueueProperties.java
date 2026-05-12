package com.example.messagequeue.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "message-queue")
public class MessageQueueProperties {

    private int maxMessageSizeBytes = 1048576;

    private int maxQueueSize = 1000;

    private int waitTimeoutSeconds = 30;

    private String offsetDir = "./data/offsets";

    private int offsetCompactionIntervalMinutes = 60;
}
