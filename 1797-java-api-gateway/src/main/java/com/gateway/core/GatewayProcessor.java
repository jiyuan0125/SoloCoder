package com.gateway.core;

import com.gateway.filter.FilterChainManager;
import com.gateway.model.FilterResult;
import com.gateway.forward.ForwardException;
import com.gateway.forward.ForwardHandler;
import com.gateway.model.BackendServer;
import com.gateway.model.RequestContext;
import com.gateway.model.RouteRule;
import com.gateway.router.RouteMatcher;
import com.gateway.router.RouteStore;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

@Slf4j
@Component
public class GatewayProcessor {

    private final RouteStore routeStore;
    private final RouteMatcher routeMatcher;
    private final FilterChainManager filterChainManager;
    private final ForwardHandler forwardHandler;
    private final String adminPrefix;

    public GatewayProcessor(
            RouteStore routeStore,
            RouteMatcher routeMatcher,
            FilterChainManager filterChainManager,
            ForwardHandler forwardHandler,
            @Value("${gateway.admin.prefix:/admin}") String adminPrefix) {
        this.routeStore = routeStore;
        this.routeMatcher = routeMatcher;
        this.filterChainManager = filterChainManager;
        this.forwardHandler = forwardHandler;
        this.adminPrefix = adminPrefix;
    }

    public void process(HttpServletRequest request, HttpServletResponse response) {
        String requestPath = request.getRequestURI();
        log.debug("处理请求: {} {}", request.getMethod(), requestPath);

        if (requestPath.startsWith(adminPrefix)) {
            return;
        }

        RequestContext context = new RequestContext();
        context.setRequest(request);
        context.setResponse(response);
        context.setRequestPath(requestPath);

        RouteRule matchedRoute = routeMatcher.matchRoute(requestPath, routeStore.getAllRoutes());
        if (matchedRoute == null) {
            throw new ForwardException("NOT_FOUND", 404, "未找到匹配的路由规则", null);
        }
        context.setMatchedRoute(matchedRoute);
        log.debug("匹配到路由: {} -> {}", requestPath, matchedRoute.getId());

        BackendServer backend = routeMatcher.selectBackend(matchedRoute.getBackends());
        if (backend == null) {
            throw new ForwardException("BAD_GATEWAY", 502, "路由 " + matchedRoute.getId() + " 没有可用的后端服务器", null);
        }
        context.setSelectedBackend(backend);
        log.debug("选择后端服务器: {}", backend.getUrl());

        FilterResult filterResult = filterChainManager.executeFilterChain(matchedRoute.getFilterChainId(), context);
        if (filterResult.getStatus() == FilterResult.Status.REJECT) {
            throw new FilterRejectException(
                    filterResult.getErrorCode(),
                    filterResult.getHttpStatus(),
                    filterResult.getErrorMessage());
        }

        forwardHandler.forward(context);
    }

    public static class FilterRejectException extends RuntimeException {
        private final String errorCode;
        private final int httpStatus;

        public FilterRejectException(String errorCode, int httpStatus, String message) {
            super(message);
            this.errorCode = errorCode;
            this.httpStatus = httpStatus;
        }

        public String getErrorCode() {
            return errorCode;
        }

        public int getHttpStatus() {
            return httpStatus;
        }
    }
}
