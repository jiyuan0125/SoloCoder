package com.example.gateway.handler;

import com.example.gateway.filter.FilterResult;
import com.example.gateway.filter.GatewayFilter;
import com.example.gateway.model.BackendTarget;
import com.example.gateway.model.RouteRule;
import com.example.gateway.service.ForwardService;
import com.example.gateway.service.RouteService;
import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Component
public class GatewayHandler implements HandlerInterceptor {

    private final RouteService routeService;
    private final ForwardService forwardService;
    private final List<GatewayFilter> filters;
    private final ObjectMapper objectMapper = new ObjectMapper();

    public GatewayHandler(RouteService routeService, ForwardService forwardService, List<GatewayFilter> filters) {
        this.routeService = routeService;
        this.forwardService = forwardService;
        this.filters = filters;
    }

    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {
        String path = request.getRequestURI();
        
        if (path.startsWith("/admin/")) {
            return true;
        }
        
        RouteRule route = routeService.matchRoute(path);
        if (route == null) {
            sendErrorResponse(response, 404, "NOT_FOUND", "未找到匹配的路由规则");
            return false;
        }
        
        BackendTarget backend = routeService.selectBackend(route);
        if (backend == null) {
            sendErrorResponse(response, 503, "SERVICE_UNAVAILABLE", "路由规则未配置后端服务");
            return false;
        }
        
        List<String> filterNames = route.getFilterNames();
        if (filterNames != null && !filterNames.isEmpty()) {
            for (String filterName : filterNames) {
                GatewayFilter filter = findFilter(filterName);
                if (filter != null) {
                    FilterResult result = filter.filter(request, response, route);
                    if (result == FilterResult.REJECT) {
                        return false;
                    }
                }
            }
        }
        
        forwardService.forward(request, response, backend);
        return false;
    }

    private GatewayFilter findFilter(String name) {
        for (GatewayFilter filter : filters) {
            if (filter.getName().equals(name)) {
                return filter;
            }
        }
        return null;
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
