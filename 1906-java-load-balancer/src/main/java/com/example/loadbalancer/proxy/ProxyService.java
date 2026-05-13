package com.example.loadbalancer.proxy;

import com.example.loadbalancer.model.Node;
import com.example.loadbalancer.strategy.StrategyManager;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.util.MultiValueMap;
import org.springframework.web.client.RestTemplate;

import javax.servlet.http.HttpServletRequest;
import java.util.Enumeration;
import java.util.HashMap;
import java.util.Map;

@Service
public class ProxyService {

    private static final Logger log = LoggerFactory.getLogger(ProxyService.class);

    private final StrategyManager strategyManager;
    private final RestTemplate restTemplate;

    public ProxyService(StrategyManager strategyManager, RestTemplate restTemplate) {
        this.strategyManager = strategyManager;
        this.restTemplate = restTemplate;
    }

    public ResponseEntity<byte[]> proxyRequest(HttpServletRequest request, byte[] body) {
        Node selectedNode = strategyManager.selectNode();
        
        if (selectedNode == null) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                    .body("No available nodes".getBytes());
        }

        selectedNode.incrementConnections();
        
        try {
            String targetUrl = buildTargetUrl(selectedNode, request);
            HttpMethod method = HttpMethod.valueOf(request.getMethod());
            HttpHeaders headers = extractHeaders(request);
            
            HttpEntity<byte[]> entity = new HttpEntity<>(body, headers);
            
            log.debug("Forwarding {} to {}", method, targetUrl);
            
            ResponseEntity<byte[]> response = restTemplate.exchange(
                    targetUrl,
                    method,
                    entity,
                    byte[].class
            );
            
            return ResponseEntity
                    .status(response.getStatusCode())
                    .headers(filterResponseHeaders(response.getHeaders()))
                    .body(response.getBody());
                    
        } catch (Exception e) {
            log.error("Error forwarding request to node {}: {}", selectedNode.getAddress(), e.getMessage());
            return ResponseEntity.status(HttpStatus.BAD_GATEWAY)
                    .body(("Error forwarding request: " + e.getMessage()).getBytes());
        } finally {
            selectedNode.decrementConnections();
        }
    }

    private String buildTargetUrl(Node node, HttpServletRequest request) {
        StringBuilder url = new StringBuilder();
        url.append("http://")
           .append(node.getAddress());
        
        if (request.getRequestURI() != null) {
            url.append(request.getRequestURI());
        }
        
        if (request.getQueryString() != null) {
            url.append("?").append(request.getQueryString());
        }
        
        return url.toString();
    }

    private HttpHeaders extractHeaders(HttpServletRequest request) {
        HttpHeaders headers = new HttpHeaders();
        Enumeration<String> headerNames = request.getHeaderNames();
        
        while (headerNames.hasMoreElements()) {
            String headerName = headerNames.nextElement();
            Enumeration<String> headerValues = request.getHeaders(headerName);
            
            while (headerValues.hasMoreElements()) {
                headers.add(headerName, headerValues.nextElement());
            }
        }
        
        headers.remove(HttpHeaders.HOST);
        headers.remove(HttpHeaders.CONTENT_LENGTH);
        
        return headers;
    }

    private HttpHeaders filterResponseHeaders(HttpHeaders responseHeaders) {
        HttpHeaders filtered = new HttpHeaders();
        
        for (Map.Entry<String, java.util.List<String>> entry : responseHeaders.entrySet()) {
            String headerName = entry.getKey();
            if (!headerName.equalsIgnoreCase(HttpHeaders.TRANSFER_ENCODING) &&
                !headerName.equalsIgnoreCase(HttpHeaders.CONNECTION)) {
                filtered.put(headerName, entry.getValue());
            }
        }
        
        return filtered;
    }
}
