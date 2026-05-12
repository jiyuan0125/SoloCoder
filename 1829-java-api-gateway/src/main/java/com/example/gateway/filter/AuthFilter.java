package com.example.gateway.filter;

import com.example.gateway.model.RouteRule;
import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.Map;

@Component
public class AuthFilter implements GatewayFilter {

    private final ObjectMapper objectMapper = new ObjectMapper();

    @Override
    public String getName() {
        return "auth";
    }

    @Override
    public FilterResult filter(HttpServletRequest request, HttpServletResponse response, RouteRule route) throws Exception {
        String authHeader = request.getHeader("Authorization");
        if (authHeader != null && authHeader.startsWith("Bearer ")) {
            String token = authHeader.substring(7);
            if (isValidToken(token)) {
                request.setAttribute("userId", extractUserId(token));
                return FilterResult.PASS;
            }
        }
        sendErrorResponse(response, 401, "UNAUTHORIZED", "认证失败：缺少或无效的认证令牌");
        return FilterResult.REJECT;
    }

    private boolean isValidToken(String token) {
        return token != null && !token.isEmpty();
    }

    private String extractUserId(String token) {
        return "user-" + token.hashCode();
    }

    private void sendErrorResponse(HttpServletResponse response, int status, String code, String message) throws Exception {
        response.setStatus(status);
        response.setContentType("application/json;charset=UTF-8");
        Map<String, String> error = new HashMap<>();
        error.put("error", code);
        error.put("message", message);
        response.getWriter().write(objectMapper.writeValueAsString(error));
    }
}
