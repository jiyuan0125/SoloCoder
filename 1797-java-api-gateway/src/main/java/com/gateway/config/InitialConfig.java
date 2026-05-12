package com.gateway.config;

import com.gateway.model.FilterChainDefinition;
import com.gateway.filter.FilterChainManager;
import com.gateway.model.BackendServer;
import com.gateway.model.RouteRule;
import com.gateway.router.RouteStore;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.stereotype.Component;

import java.util.Arrays;
import java.util.Collections;

@Slf4j
@Component
public class InitialConfig implements ApplicationRunner {

    private final RouteStore routeStore;
    private final FilterChainManager filterChainManager;

    public InitialConfig(RouteStore routeStore, FilterChainManager filterChainManager) {
        this.routeStore = routeStore;
        this.filterChainManager = filterChainManager;
    }

    @Override
    public void run(ApplicationArguments args) {
        log.info("初始化网关配置...");

        FilterChainDefinition publicChain = new FilterChainDefinition(
                "public-chain",
                Collections.emptyList()
        );
        filterChainManager.updateFilterChain(publicChain);
        log.info("创建过滤器链: public-chain (空链)");

        FilterChainDefinition fullChain = new FilterChainDefinition(
                "full-chain",
                Arrays.asList("auth", "permission", "ratelimit")
        );
        filterChainManager.updateFilterChain(fullChain);
        log.info("创建过滤器链: full-chain (auth -> permission -> ratelimit)");

        RouteRule publicRoute = new RouteRule(
                "public-route",
                "/api/public/",
                Arrays.asList(new BackendServer("http://localhost:8081", 1)),
                "public-chain"
        );
        routeStore.addRoute(publicRoute);
        log.info("创建路由: /api/public/* -> public-chain");

        RouteRule adminRoute = new RouteRule(
                "admin-route",
                "/api/admin/",
                Arrays.asList(new BackendServer("http://localhost:8082", 1)),
                "full-chain"
        );
        routeStore.addRoute(adminRoute);
        log.info("创建路由: /api/admin/* -> full-chain");

        RouteRule userRoute = new RouteRule(
                "user-route",
                "/api/v1/user/",
                Arrays.asList(new BackendServer("http://localhost:8083", 1)),
                "full-chain"
        );
        routeStore.addRoute(userRoute);
        log.info("创建路由: /api/v1/user/* -> full-chain");

        RouteRule apiV1Route = new RouteRule(
                "api-v1-route",
                "/api/v1/",
                Arrays.asList(
                        new BackendServer("http://localhost:8084", 7),
                        new BackendServer("http://localhost:8085", 3)
                ),
                "public-chain"
        );
        routeStore.addRoute(apiV1Route);
        log.info("创建路由: /api/v1/* (权重 7:3) -> public-chain");

        log.info("网关配置初始化完成");
    }
}
