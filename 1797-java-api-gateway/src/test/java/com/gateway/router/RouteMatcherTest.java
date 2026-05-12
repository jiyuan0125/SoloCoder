package com.gateway.router;

import com.gateway.model.BackendServer;
import com.gateway.model.RouteRule;
import org.junit.jupiter.api.Test;

import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

class RouteMatcherTest {

    private final RouteMatcher routeMatcher = new RouteMatcher();

    @Test
    void testLongestPrefixMatch() {
        RouteRule apiV1 = new RouteRule("api-v1", "/api/v1/", 
                Arrays.asList(new BackendServer("http://localhost:8084", 1)), "chain1");
        RouteRule apiV1User = new RouteRule("api-v1-user", "/api/v1/user/", 
                Arrays.asList(new BackendServer("http://localhost:8083", 1)), "chain2");
        RouteRule apiPublic = new RouteRule("api-public", "/api/public/", 
                Arrays.asList(new BackendServer("http://localhost:8081", 1)), "chain3");
        
        List<RouteRule> routes = Arrays.asList(apiV1, apiV1User, apiPublic);

        RouteRule matched1 = routeMatcher.matchRoute("/api/v1/user/list", routes);
        assertEquals("api-v1-user", matched1.getId());

        RouteRule matched2 = routeMatcher.matchRoute("/api/v1/product/list", routes);
        assertEquals("api-v1", matched2.getId());

        RouteRule matched3 = routeMatcher.matchRoute("/api/public/health", routes);
        assertEquals("api-public", matched3.getId());
    }

    @Test
    void testWeightedRouting() {
        BackendServer b1 = new BackendServer("http://server1", 7);
        BackendServer b2 = new BackendServer("http://server2", 3);
        List<BackendServer> backends = Arrays.asList(b1, b2);

        Map<String, Integer> counts = new HashMap<>();
        for (int i = 0; i < 1000; i++) {
            BackendServer selected = routeMatcher.selectBackend(backends);
            counts.merge(selected.getUrl(), 1, Integer::sum);
        }

        int count1 = counts.getOrDefault("http://server1", 0);
        int count2 = counts.getOrDefault("http://server2", 0);

        assertTrue(count1 > count2);
        assertTrue(count1 >= 600);
        assertTrue(count2 >= 200);
    }

    @Test
    void testNoMatchReturnsNull() {
        RouteRule apiV1 = new RouteRule("api-v1", "/api/v1/", 
                Arrays.asList(new BackendServer("http://localhost:8084", 1)), "chain1");
        List<RouteRule> routes = Arrays.asList(apiV1);

        RouteRule matched = routeMatcher.matchRoute("/other/path", routes);
        assertNull(matched);
    }
}
