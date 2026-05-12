package com.trace.servicea.config;

import org.springframework.web.service.annotation.GetExchange;

public interface ServiceBClient {

    @GetExchange("/api/b")
    String callServiceB();

}
