package com.canary.gateway.controller;

import com.canary.gateway.service.CanaryService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.*;
import org.springframework.util.StreamUtils;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.client.HttpStatusCodeException;
import org.springframework.web.client.RestTemplate;

import javax.servlet.http.HttpServletRequest;
import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.Enumeration;

@RestController
public class ProxyController {

    @Autowired
    private CanaryService canaryService;

    private final RestTemplate restTemplate = new RestTemplate();

    @RequestMapping(value = "/**", method = {RequestMethod.GET, RequestMethod.POST, RequestMethod.PUT, RequestMethod.DELETE, RequestMethod.PATCH})
    public ResponseEntity<?> proxy(HttpServletRequest request,
                                    @RequestHeader(value = "X-User-ID", required = false) String userId,
                                    @RequestBody(required = false) byte[] body) {

        if (userId == null || userId.isEmpty()) {
            return ResponseEntity.badRequest().body("X-User-ID header is required");
        }

        String targetBackend = canaryService.getTargetBackend(userId);
        boolean isGray = targetBackend.equals(canaryService.getTargetBackend(userId)) && canaryService.shouldRouteToGray(userId);

        try {
            String url = buildUrl(request, targetBackend);
            
            HttpHeaders headers = extractHeaders(request);
            HttpEntity<byte[]> entity = new HttpEntity<>(body, headers);
            
            ResponseEntity<byte[]> response = restTemplate.exchange(
                    url,
                    HttpMethod.valueOf(request.getMethod()),
                    entity,
                    byte[].class
            );

            if (isGray) {
                canaryService.incrementGrayStats();
            } else {
                canaryService.incrementProdStats();
            }

            return ResponseEntity.status(response.getStatusCode())
                    .headers(response.getHeaders())
                    .body(response.getBody());

        } catch (HttpStatusCodeException e) {
            if (isGray) {
                canaryService.incrementGrayStats();
            } else {
                canaryService.incrementProdStats();
            }
            return ResponseEntity.status(e.getStatusCode())
                    .headers(e.getResponseHeaders())
                    .body(e.getResponseBodyAsByteArray());
        } catch (Exception e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(("Proxy error: " + e.getMessage()).getBytes());
        }
    }

    private String buildUrl(HttpServletRequest request, String backend) throws URISyntaxException {
        String path = request.getRequestURI();
        String query = request.getQueryString();
        String url = backend + path;
        if (query != null && !query.isEmpty()) {
            url += "?" + query;
        }
        return url;
    }

    private HttpHeaders extractHeaders(HttpServletRequest request) {
        HttpHeaders headers = new HttpHeaders();
        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String headerName = headerNames.nextElement();
            if (!headerName.equalsIgnoreCase("host") && 
                !headerName.equalsIgnoreCase("connection") &&
                !headerName.equalsIgnoreCase("content-length")) {
                Enumeration<String> headerValues = request.getHeaders(headerName);
                while (headerValues.hasMoreElements()) {
                    headers.add(headerName, headerValues.nextElement());
                }
            }
        }
        return headers;
    }
}
