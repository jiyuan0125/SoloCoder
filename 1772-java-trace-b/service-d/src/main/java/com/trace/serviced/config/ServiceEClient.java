package com.trace.serviced.config;

import org.springframework.web.service.annotation.GetExchange;

public interface ServiceEClient {

    @GetExchange("/api/e")
    String callServiceE();

}
