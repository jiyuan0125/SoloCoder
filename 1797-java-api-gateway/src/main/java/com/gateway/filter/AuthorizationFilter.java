package com.gateway.filter;

import com.gateway.model.FilterResult;
import com.gateway.model.RequestContext;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import javax.servlet.http.HttpServletRequest;
import java.util.HashMap;
import java.util.Map;

@Slf4j
@Component
public class AuthorizationFilter implements GatewayFilter {

    public static final String NAME = "permission";

    private final Map<String, String> pathRoleMap = new HashMap<>();

    public AuthorizationFilter() {
        pathRoleMap.put("/admin/", "ADMIN");
        pathRoleMap.put("/api/admin/", "ADMIN");
        pathRoleMap.put("/api/user/", "USER");
    }

    @Override
    public String getName() {
        return NAME;
    }

    @Override
    public FilterResult filter(RequestContext context) {
        HttpServletRequest request = context.getRequest();
        String requestPath = request.getRequestURI();
        String userId = context.getAttribute("userId");

        if (userId == null) {
            return FilterResult.reject("FORBIDDEN", "用户未认证", 403);
        }

        String requiredRole = null;
        for (Map.Entry<String, String> entry : pathRoleMap.entrySet()) {
            if (requestPath.startsWith(entry.getKey())) {
                requiredRole = entry.getValue();
                break;
            }
        }

        if (requiredRole != null) {
            String userRole = request.getHeader("X-User-Role");
            if (userRole == null || !userRole.contains(requiredRole)) {
                return FilterResult.reject("FORBIDDEN", "权限不足，需要角色: " + requiredRole, 403);
            }
        }

        return FilterResult.pass();
    }
}
