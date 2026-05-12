package com.gateway.core;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

@Slf4j
@Component
public class GatewayServlet implements HandlerInterceptor {

    private final GatewayProcessor gatewayProcessor;

    public GatewayServlet(GatewayProcessor gatewayProcessor) {
        this.gatewayProcessor = gatewayProcessor;
    }

    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {
        gatewayProcessor.process(request, response);
        return false;
    }
}
