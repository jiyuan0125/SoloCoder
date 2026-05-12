package com.trace.serviceb.config;

import org.springframework.web.service.annotation.GetExchange;

public interface ServiceCClient {

    @GetExchange("/api/c")
    String callServiceC();

}
