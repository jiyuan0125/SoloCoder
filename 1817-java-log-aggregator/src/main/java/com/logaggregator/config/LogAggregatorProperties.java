package com.logaggregator.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "log-aggregator")
public class LogAggregatorProperties {

    private int retentionDays = 7;
    private int maxMessageSize = 10240;
    private int defaultPageSize = 50;
}
