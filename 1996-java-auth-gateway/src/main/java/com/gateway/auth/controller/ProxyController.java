package com.gateway.auth.controller;

import com.gateway.auth.config.RouteConfig;
import com.gateway.auth.model.AuditLog;
import com.gateway.auth.service.AuditLogService;
import org.springframework.http.*;
import org.springframework.util.StreamUtils;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.HttpClientErrorException;
import org.springframework.web.client.HttpServerErrorException;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.client.RestTemplate;

import javax.servlet.http.HttpServletRequest;
import java.io.IOException;
import java.net.URI;
import java.time.Instant;
import java.util.Enumeration;
import java.util.HashMap;
import java.util.Map;

@RestController
public class ProxyController {

    private final RouteConfig routeConfig;
    private final AuditLogService auditLogService;
    private final RestTemplate restTemplate = new RestTemplate();

    public ProxyController(RouteConfig routeConfig, AuditLogService auditLogService) {
        this.routeConfig = routeConfig;
        this.auditLogService = auditLogService;
    }

    @RequestMapping("/**")
    public ResponseEntity<byte[]> proxy(HttpServletRequest request) throws IOException {
        String path = request.getRequestURI();
        String queryString = request.getQueryString();
        String clientId = (String) request.getAttribute("clientId");

        RouteConfig.Route matchedRoute = null;
        for (RouteConfig.Route route : routeConfig.getRoutes()) {
            if (path.startsWith(route.getPathPrefix())) {
                matchedRoute = route;
                break;
            }
        }

        if (matchedRoute == null) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(("No route configured for path: " + path).getBytes());
        }

        String remainingPath = path.substring(matchedRoute.getPathPrefix().length());
        if (!remainingPath.startsWith("/")) {
            remainingPath = "/" + remainingPath;
        }

        String targetUrl = matchedRoute.getUrl() + remainingPath;
        if (queryString != null && !queryString.isEmpty()) {
            targetUrl += "?" + queryString;
        }

        HttpMethod method = HttpMethod.valueOf(request.getMethod());
        HttpHeaders headers = buildHeaders(request);
        byte[] body = StreamUtils.copyToByteArray(request.getInputStream());
        HttpEntity<byte[]> requestEntity = new HttpEntity<>(body, headers);

        try {
            ResponseEntity<byte[]> response = restTemplate.exchange(
                    URI.create(targetUrl),
                    method,
                    requestEntity,
                    byte[].class
            );

            if (response.getStatusCode().is4xxClientError()) {
                auditLogService.log(new AuditLog(Instant.now(), clientId, path, AuditLog.Result.UPSTREAM_4XX));
            }

            HttpHeaders responseHeaders = new HttpHeaders();
            response.getHeaders().forEach((key, value) -> {
                if (!key.equalsIgnoreCase(HttpHeaders.TRANSFER_ENCODING) &&
                    !key.equalsIgnoreCase(HttpHeaders.CONNECTION) &&
                    !key.equalsIgnoreCase(HttpHeaders.CONTENT_LENGTH)) {
                    responseHeaders.put(key, value);
                }
            });

            return ResponseEntity.status(response.getStatusCode())
                    .headers(responseHeaders)
                    .body(response.getBody());

        } catch (HttpClientErrorException e) {
            auditLogService.log(new AuditLog(Instant.now(), clientId, path, AuditLog.Result.UPSTREAM_4XX));

            HttpHeaders responseHeaders = new HttpHeaders();
            e.getResponseHeaders().forEach((key, value) -> {
                if (!key.equalsIgnoreCase(HttpHeaders.TRANSFER_ENCODING) &&
                    !key.equalsIgnoreCase(HttpHeaders.CONNECTION)) {
                    responseHeaders.put(key, value);
                }
            });

            return ResponseEntity.status(e.getStatusCode())
                    .headers(responseHeaders)
                    .body(e.getResponseBodyAsByteArray());

        } catch (HttpServerErrorException e) {
            return ResponseEntity.status(e.getStatusCode())
                    .body(e.getResponseBodyAsByteArray());

        } catch (ResourceAccessException e) {
            return ResponseEntity.status(HttpStatus.BAD_GATEWAY)
                    .body(("Bad Gateway: " + e.getMessage()).getBytes());
        }
    }

    private HttpHeaders buildHeaders(HttpServletRequest request) {
        HttpHeaders headers = new HttpHeaders();
        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String name = headerNames.nextElement();
            Enumeration<String> values = request.getHeaders(name);
            while (values.hasMoreElements()) {
                headers.add(name, values.nextElement());
            }
        }
        return headers;
    }
}
