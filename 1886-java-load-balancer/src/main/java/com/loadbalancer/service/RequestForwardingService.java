package com.loadbalancer.service;

import com.loadbalancer.model.Node;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.*;
import org.springframework.stereotype.Service;

import java.io.*;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.Enumeration;
import java.util.List;

@Slf4j
@Service
@RequiredArgsConstructor
public class RequestForwardingService {
    private final LoadBalancerService loadBalancerService;
    private final HttpClient httpClient = HttpClient.newHttpClient();

    public void forwardRequest(HttpServletRequest request, HttpServletResponse response) throws IOException {
        Node targetNode = loadBalancerService.selectNode();
        if (targetNode == null) {
            response.setStatus(HttpStatus.SERVICE_UNAVAILABLE.value());
            response.getWriter().write("No available backend nodes");
            return;
        }

        targetNode.incrementActiveConnections();
        try {
            String targetUrl = buildTargetUrl(request, targetNode);
            HttpRequest.Builder requestBuilder = HttpRequest.newBuilder()
                    .uri(URI.create(targetUrl))
                    .method(request.getMethod(), HttpRequest.BodyPublishers.noBody());

            copyHeaders(request, requestBuilder);

            HttpRequest httpRequest = requestBuilder.build();
            
            HttpResponse<byte[]> httpResponse = httpClient.send(httpRequest, HttpResponse.BodyHandlers.ofByteArray());
            
            response.setStatus(httpResponse.statusCode());
            
            httpResponse.headers().map().forEach((key, values) -> {
                if (!isHopByHopHeader(key)) {
                    for (String value : values) {
                        response.addHeader(key, value);
                    }
                }
            });

            byte[] responseBody = httpResponse.body();
            if (responseBody != null && responseBody.length > 0) {
                response.getOutputStream().write(responseBody);
            }
        } catch (Exception e) {
            log.error("Error forwarding request to {}: {}", targetNode.getAddress(), e.getMessage());
            response.setStatus(HttpStatus.BAD_GATEWAY.value());
            response.getWriter().write("Error forwarding request: " + e.getMessage());
        } finally {
            targetNode.decrementActiveConnections();
        }
    }

    private String buildTargetUrl(HttpServletRequest request, Node targetNode) {
        StringBuilder url = new StringBuilder();
        url.append("http://").append(targetNode.getAddress());
        if (request.getRequestURI() != null) {
            url.append(request.getRequestURI());
        }
        if (request.getQueryString() != null) {
            url.append("?").append(request.getQueryString());
        }
        return url.toString();
    }

    private void copyHeaders(HttpServletRequest request, HttpRequest.Builder requestBuilder) {
        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String headerName = headerNames.nextElement();
            if (isHopByHopHeader(headerName)) {
                continue;
            }
            Enumeration<String> headerValues = request.getHeaders(headerName);
            while (headerValues.hasMoreElements()) {
                String headerValue = headerValues.nextElement();
                requestBuilder.header(headerName, headerValue);
            }
        }
    }

    private boolean isHopByHopHeader(String headerName) {
        List<String> hopByHopHeaders = List.of(
                "connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
                "te", "trailers", "transfer-encoding", "upgrade", "host"
        );
        return hopByHopHeaders.contains(headerName.toLowerCase());
    }
}
