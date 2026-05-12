package com.loadbalancer.controller;

import com.loadbalancer.model.BackendNode;
import com.loadbalancer.service.HealthCheckService;
import com.loadbalancer.service.LoadBalancerService;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.client.HttpStatusCodeException;
import org.springframework.web.client.RestTemplate;

import java.util.Optional;

@RestController
public class ProxyController {
    private final LoadBalancerService loadBalancerService;
    private final HealthCheckService healthCheckService;
    private final RestTemplate restTemplate = new RestTemplate();

    public ProxyController(LoadBalancerService loadBalancerService,
                         HealthCheckService healthCheckService) {
        this.loadBalancerService = loadBalancerService;
        this.healthCheckService = healthCheckService;
    }

    @RequestMapping("/**")
    public ResponseEntity<byte[]> proxyRequest(HttpServletRequest request,
                                         @RequestParam(required = false) java.util.Map<String, String> params,
                                         @RequestBody(required = false) byte[] body) {
        
        String requestPath = request.getRequestURI();
        
        Optional<BackendNode> selectedNode = loadBalancerService.selectBackend(requestPath);
        
        if (selectedNode.isEmpty()) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                    .body(("No healthy backend nodes available for path: " + requestPath).getBytes());
        }

        BackendNode node = selectedNode.get();
        String targetUrl = buildTargetUrl(node, requestPath, request.getQueryString());

        try {
            HttpMethod method = HttpMethod.valueOf(request.getMethod());
            HttpHeaders headers = extractHeaders(request);
            
            org.springframework.http.HttpEntity<byte[]> entity = new org.springframework.http.HttpEntity<>(body, headers);
            
            ResponseEntity<byte[]> response = restTemplate.exchange(
                    targetUrl,
                    method,
                    entity,
                    byte[].class);

            healthCheckService.recordForwardingSuccess(node.getId());
            return response;

        } catch (HttpStatusCodeException e) {
            healthCheckService.recordForwardingFailure(node.getId());
            return ResponseEntity.status(e.getStatusCode())
                    .headers(e.getResponseHeaders())
                    .body(e.getResponseBodyAsByteArray());
        } catch (Exception e) {
            healthCheckService.recordForwardingFailure(node.getId());
            return ResponseEntity.status(HttpStatus.BAD_GATEWAY)
                    .body(("Error forwarding to backend: " + e.getMessage()).getBytes());
        }
    }

    private String buildTargetUrl(BackendNode node, String path, String queryString) {
        StringBuilder url = new StringBuilder("http://");
        url.append(node.getHost()).append(":").append(node.getPort());
        url.append(path);
        if (queryString != null && !queryString.isEmpty()) {
            url.append("?").append(queryString);
        }
        return url.toString();
    }

    private HttpHeaders extractHeaders(HttpServletRequest request) {
        HttpHeaders headers = new HttpHeaders();
        java.util.Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String headerName = headerNames.nextElement();
            if (!headerName.equalsIgnoreCase("host")) {
                headers.set(headerName, request.getHeader(headerName));
            }
        }
        return headers;
    }
}
