package com.gateway.filter;

import com.gateway.model.FilterResult;
import com.gateway.model.RequestContext;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class AuthenticationFilter implements GatewayFilter {

    public static final String NAME = "auth";

    @Override
    public String getName() {
        return NAME;
    }

    @Override
    public FilterResult filter(RequestContext context) {
        String authHeader = context.getRequest().getHeader("Authorization");
        if (authHeader == null || authHeader.isEmpty()) {
            return FilterResult.reject("UNAUTHORIZED", "缺少认证令牌", 401);
        }

        if (!authHeader.startsWith("Bearer ")) {
            return FilterResult.reject("UNAUTHORIZED", "无效的认证格式", 401);
        }

        String token = authHeader.substring(7);
        if ("invalid".equals(token)) {
            return FilterResult.reject("UNAUTHORIZED", "认证令牌无效", 401);
        }

        context.setAttribute("userId", "user-" + token.hashCode());
        return FilterResult.pass();
    }
}
