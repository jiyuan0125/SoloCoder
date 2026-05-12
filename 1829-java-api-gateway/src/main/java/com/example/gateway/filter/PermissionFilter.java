package com.example.gateway.filter;

import com.example.gateway.model.RouteRule;
import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;

import java.util.HashMap;
import java.util.Map;

@Component
public class PermissionFilter implements GatewayFilter {

    private final ObjectMapper objectMapper = new ObjectMapper();

    @Override
    public String getName() {
        return "permission";
    }

    @Override
    public FilterResult filter(HttpServletRequest request, HttpServletResponse response, RouteRule route) throws Exception {
        String path = request.getRequestURI();
        if (path.startsWith("/api/admin/")) {
            String userId = (String) request.getAttribute("userId");
            if (!isAdmin(userId)) {
                sendErrorResponse(response, 403, "FORBIDDEN", "权限不足：需要管理员权限");
                return FilterResult.REJECT;
            }
        }
        return FilterResult.PASS;
    }

    private boolean isAdmin(String userId) {
        return userId != null && userId.contains("admin");
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
