package com.example.gateway.filter;

import com.example.gateway.model.RouteRule;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;

public interface GatewayFilter {
    String getName();
    FilterResult filter(HttpServletRequest request, HttpServletResponse response, RouteRule route) throws Exception;
}
