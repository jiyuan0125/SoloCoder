package com.trace.gateway.config;

import org.springframework.web.service.annotation.GetExchange;

public interface ServiceAClient {

    @GetExchange("/api/a")
    String callServiceA();

}
