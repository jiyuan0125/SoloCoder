package com.loadbalancer.controller;

import com.loadbalancer.model.BackendNode;
import com.loadbalancer.service.HealthCheckService;
import com.loadbalancer.service.LoadBalancingService;
import org.springframework.http.*;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.client.HttpClientErrorException;
import org.springframework.web.client.HttpServerErrorException;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.client.RestTemplate;

import javax.servlet.http.HttpServletRequest;
import java.net.URI;
import java.util.Enumeration;
import java.util.List;

@RestController
public class LoadBalancingController {
    private final LoadBalancingService loadBalancingService;
    private final HealthCheckService healthCheckService;
    private final RestTemplate restTemplate = new RestTemplate();

    public LoadBalancingController(LoadBalancingService loadBalancingService,
                                    HealthCheckService healthCheckService) {
        this.loadBalancingService = loadBalancingService;
        this.healthCheckService = healthCheckService;
    }

    @RequestMapping(value = "/**", method = {RequestMethod.GET, RequestMethod.POST, RequestMethod.PUT,
                                              RequestMethod.DELETE, RequestMethod.PATCH, RequestMethod.OPTIONS})
    public ResponseEntity<?> forward(HttpServletRequest request,
                                      @RequestBody(required = false) byte[] body) {
        String path = request.getRequestURI();

        if (path.startsWith("/lb-api/")) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND).body("Not found");
        }

        BackendNode node = loadBalancingService.selectNode(path);
        if (node == null) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                    .body("No available backend nodes");
        }

        try {
            String targetUrl = buildTargetUrl(node, request);
            HttpMethod method = HttpMethod.valueOf(request.getMethod());
            HttpHeaders headers = extractHeaders(request);

            HttpEntity<byte[]> entity = new HttpEntity<>(body, headers);
            ResponseEntity<byte[]> response = restTemplate.exchange(
                    targetUrl, method, entity, byte[].class);

            return ResponseEntity.status(response.getStatusCode())
                    .headers(response.getHeaders())
                    .body(response.getBody());

        } catch (ResourceAccessException | HttpClientErrorException | HttpServerErrorException e) {
            healthCheckService.recordForwardFailure(node.getId());
            return ResponseEntity.status(HttpStatus.BAD_GATEWAY)
                    .body("Failed to forward to backend: " + e.getMessage());
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body("Load balancer error: " + e.getMessage());
        }
    }

    private String buildTargetUrl(BackendNode node, HttpServletRequest request) {
        String scheme = request.isSecure() ? "https" : "http";
        String queryString = request.getQueryString();
        String path = request.getRequestURI();
        if (queryString != null && !queryString.isEmpty()) {
            path = path + "?" + queryString;
        }
        return scheme + "://" + node.getHost() + ":" + node.getPort() + path;
    }

    private HttpHeaders extractHeaders(HttpServletRequest request) {
        HttpHeaders headers = new HttpHeaders();
        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String name = headerNames.nextElement();
            if (shouldForwardHeader(name)) {
                List<String> values = java.util.Collections.list(request.getHeaders(name));
                headers.put(name, values);
            }
        }
        return headers;
    }

    private boolean shouldForwardHeader(String name) {
        String lower = name.toLowerCase();
        return !lower.equals("host") &&
               !lower.equals("content-length") &&
               !lower.equals("connection");
    }
}
