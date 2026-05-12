package com.gateway.router;

import com.gateway.model.BackendServer;
import com.gateway.model.RouteRule;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Random;
import java.util.concurrent.atomic.AtomicInteger;

@Slf4j
@Component
public class RouteMatcher {

    private final Random random = new Random();
    private final AtomicInteger counter = new AtomicInteger(0);

    public RouteRule matchRoute(String requestPath, List<RouteRule> routes) {
        if (routes == null || routes.isEmpty()) {
            return null;
        }

        RouteRule bestMatch = null;
        int maxLength = -1;

        for (RouteRule route : routes) {
            String prefix = normalizePrefix(route.getPathPrefix());
            if (isPrefixMatch(requestPath, prefix)) {
                if (prefix.length() > maxLength) {
                    maxLength = prefix.length();
                    bestMatch = route;
                }
            }
        }

        return bestMatch;
    }

    private String normalizePrefix(String prefix) {
        if (prefix == null) {
            return "";
        }
        if (!prefix.startsWith("/")) {
            prefix = "/" + prefix;
        }
        if (!prefix.endsWith("/")) {
            prefix = prefix + "/";
        }
        return prefix;
    }

    private boolean isPrefixMatch(String requestPath, String prefix) {
        if (requestPath == null) {
            return false;
        }
        if (!requestPath.startsWith("/")) {
            requestPath = "/" + requestPath;
        }
        if (!requestPath.endsWith("/")) {
            requestPath = requestPath + "/";
        }
        return requestPath.startsWith(prefix);
    }

    public BackendServer selectBackend(List<BackendServer> backends) {
        if (backends == null || backends.isEmpty()) {
            return null;
        }

        if (backends.size() == 1) {
            return backends.get(0);
        }

        int totalWeight = backends.stream()
                .mapToInt(BackendServer::getWeight)
                .sum();

        if (totalWeight <= 0) {
            int index = Math.abs(counter.incrementAndGet()) % backends.size();
            return backends.get(index);
        }

        int randomWeight = random.nextInt(totalWeight);
        int currentWeight = 0;

        for (BackendServer backend : backends) {
            currentWeight += backend.getWeight();
            if (randomWeight < currentWeight) {
                return backend;
            }
        }

        return backends.get(backends.size() - 1);
    }
}
