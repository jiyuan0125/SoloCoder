package com.example.apiproxymirror.controller;

import com.example.apiproxymirror.config.ProxyConfig;
import com.example.apiproxymirror.filter.CachedBodyHttpServletRequest;
import com.example.apiproxymirror.filter.CachedBodyHttpServletResponse;
import com.example.apiproxymirror.model.RecordedExchange;
import com.example.apiproxymirror.service.RecordingService;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import lombok.SneakyThrows;
import org.springframework.http.*;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.client.HttpStatusCodeException;
import org.springframework.web.client.RestTemplate;

import java.io.IOException;
import java.util.Enumeration;
import java.util.HashMap;
import java.util.Map;

@Controller
@RequiredArgsConstructor
public class ProxyController {

    private final RestTemplate restTemplate;
    private final ProxyConfig proxyConfig;
    private final RecordingService recordingService;

    private static final String X_RECORDED_HEADER = "X-Recorded";

    @RequestMapping("/**")
    @SneakyThrows
    public void proxy(HttpServletRequest request, HttpServletResponse response) {
        String path = request.getRequestURI();

        if (shouldBypassProxy(path)) {
            return;
        }

        CachedBodyHttpServletRequest cachedRequest = new CachedBodyHttpServletRequest(request);
        CachedBodyHttpServletResponse cachedResponse = new CachedBodyHttpServletResponse(response);

        boolean shouldRecord = recordingService.shouldRecord(path);

        try {
            ResponseEntity<String> backendResponse = forwardRequest(cachedRequest, shouldRecord);
            copyResponse(backendResponse, cachedResponse);

            if (shouldRecord) {
                recordExchange(cachedRequest, cachedResponse, backendResponse);
            }
        } catch (HttpStatusCodeException e) {
            handleException(e, cachedResponse);

            if (shouldRecord) {
                recordExchange(cachedRequest, cachedResponse, e);
            }
        }
    }

    private boolean shouldBypassProxy(String path) {
        return path.startsWith("/mirror/") ||
                path.startsWith("/replay/") ||
                path.startsWith("/records") ||
                path.equals("/replay");
    }

    private ResponseEntity<String> forwardRequest(CachedBodyHttpServletRequest request, boolean addRecordedHeader) {
        String targetUrl = buildTargetUrl(request);
        HttpMethod method = HttpMethod.valueOf(request.getMethod());
        HttpHeaders headers = extractHeaders(request);

        if (addRecordedHeader) {
            headers.set(X_RECORDED_HEADER, "true");
        }

        String body = request.getCachedBodyAsString();
        HttpEntity<String> requestEntity = body.isEmpty()
                ? new HttpEntity<>(headers)
                : new HttpEntity<>(body, headers);

        return restTemplate.exchange(targetUrl, method, requestEntity, String.class);
    }

    private String buildTargetUrl(HttpServletRequest request) {
        StringBuilder url = new StringBuilder(proxyConfig.getTargetUrl());
        url.append(request.getRequestURI());
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
            if (!"content-length".equalsIgnoreCase(headerName) && !"host".equalsIgnoreCase(headerName)) {
                headers.set(headerName, request.getHeader(headerName));
            }
        }
        return headers;
    }

    private void copyResponse(ResponseEntity<String> backendResponse, CachedBodyHttpServletResponse response) throws IOException {
        response.setStatus(backendResponse.getStatusCode().value());

        backendResponse.getHeaders().forEach((name, values) -> {
            if (!"transfer-encoding".equalsIgnoreCase(name)) {
                for (String value : values) {
                    response.addHeader(name, value);
                }
            }
        });

        if (backendResponse.getBody() != null) {
            response.getWriter().write(backendResponse.getBody());
        }
    }

    private void handleException(HttpStatusCodeException e, CachedBodyHttpServletResponse response) throws IOException {
        response.setStatus(e.getStatusCode().value());

        e.getResponseHeaders().forEach((name, values) -> {
            if (!"transfer-encoding".equalsIgnoreCase(name)) {
                for (String value : values) {
                    response.addHeader(name, value);
                }
            }
        });

        if (e.getResponseBodyAsString() != null) {
            response.getWriter().write(e.getResponseBodyAsString());
        }
    }

    private void recordExchange(CachedBodyHttpServletRequest request,
                                CachedBodyHttpServletResponse response,
                                ResponseEntity<String> backendResponse) {
        Map<String, String> requestHeaders = convertToMap(extractHeaders(request));
        Map<String, String> responseHeaders = convertToMap(backendResponse.getHeaders());

        RecordedExchange exchange = RecordedExchange.builder()
                .method(request.getMethod())
                .path(request.getRequestURI())
                .requestHeaders(requestHeaders)
                .requestBody(request.getCachedBodyAsString())
                .responseStatus(backendResponse.getStatusCode().value())
                .responseHeaders(responseHeaders)
                .responseBody(backendResponse.getBody())
                .build();

        recordingService.record(exchange);
    }

    private void recordExchange(CachedBodyHttpServletRequest request,
                                CachedBodyHttpServletResponse response,
                                HttpStatusCodeException e) {
        Map<String, String> requestHeaders = convertToMap(extractHeaders(request));
        Map<String, String> responseHeaders = e.getResponseHeaders() != null
                ? convertToMap(e.getResponseHeaders())
                : new HashMap<>();

        RecordedExchange exchange = RecordedExchange.builder()
                .method(request.getMethod())
                .path(request.getRequestURI())
                .requestHeaders(requestHeaders)
                .requestBody(request.getCachedBodyAsString())
                .responseStatus(e.getStatusCode().value())
                .responseHeaders(responseHeaders)
                .responseBody(e.getResponseBodyAsString())
                .build();

        recordingService.record(exchange);
    }

    private Map<String, String> convertToMap(HttpHeaders headers) {
        Map<String, String> map = new HashMap<>();
        headers.forEach((name, values) -> {
            if (!values.isEmpty()) {
                map.put(name, values.get(0));
            }
        });
        return map;
    }
}
