package com.gateway.model;

import lombok.Data;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.util.HashMap;
import java.util.Map;

@Data
public class RequestContext {

    private HttpServletRequest request;
    private HttpServletResponse response;
    private RouteRule matchedRoute;
    private BackendServer selectedBackend;
    private String requestPath;
    private Map<String, Object> attributes = new HashMap<>();

    public void setAttribute(String key, Object value) {
        attributes.put(key, value);
    }

    @SuppressWarnings("unchecked")
    public <T> T getAttribute(String key) {
        return (T) attributes.get(key);
    }
}
