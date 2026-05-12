package com.trace.servicec.config;

import org.springframework.web.service.annotation.GetExchange;

public interface ServiceDClient {

    @GetExchange("/api/d")
    String callServiceD();

}
