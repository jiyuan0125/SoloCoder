package com.example.apiproxymirror.controller;

import com.example.apiproxymirror.config.ProxyConfig;
import com.example.apiproxymirror.filter.CachedBodyHttpServletRequest;
import com.example.apiproxymirror.filter.CachedBodyHttpServletResponse;
import com.example.apiproxymirror.model.RecordedExchange;
import com.example.apiproxymirror.service.RecordingService;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.Builder;
import lombok.Data;
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

    @Data
    @Builder
    private static class ForwardResult {
        private ResponseEntity<String> response;
        private HttpHeaders requestHeaders;
    }

    @Data
    @Builder
    private static class ErrorResult {
        private HttpStatusCodeException exception;
        private HttpHeaders requestHeaders;
    }

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
            ForwardResult result = forwardRequest(cachedRequest, shouldRecord);
            copyResponse(result.getResponse(), cachedResponse);

            if (shouldRecord) {
                recordExchange(cachedRequest, cachedResponse, result);
            }
        } catch (HttpStatusCodeException e) {
            ErrorResult errorResult = ErrorResult.builder()
                    .exception(e)
                    .requestHeaders(extractHeadersWithRecordedFlag(cachedRequest, shouldRecord))
                    .build();
            handleException(errorResult, cachedResponse);

            if (shouldRecord) {
                recordExchange(cachedRequest, cachedResponse, errorResult);
            }
        }
    }

    private boolean shouldBypassProxy(String path) {
        return path.startsWith("/mirror/") ||
                path.startsWith("/replay/") ||
                path.startsWith("/records") ||
                path.equals("/replay");
    }

    private HttpHeaders extractHeadersWithRecordedFlag(CachedBodyHttpServletRequest request, boolean addRecordedHeader) {
        HttpHeaders headers = extractHeaders(request);
        if (addRecordedHeader) {
            headers.set(X_RECORDED_HEADER, "true");
        }
        return headers;
    }

    private ForwardResult forwardRequest(CachedBodyHttpServletRequest request, boolean addRecordedHeader) {
        String targetUrl = buildTargetUrl(request);
        HttpMethod method = HttpMethod.valueOf(request.getMethod());
        HttpHeaders headers = extractHeadersWithRecordedFlag(request, addRecordedHeader);

        String body = request.getCachedBodyAsString();
        HttpEntity<String> requestEntity = body.isEmpty()
                ? new HttpEntity<>(headers)
                : new HttpEntity<>(body, headers);

        ResponseEntity<String> response = restTemplate.exchange(targetUrl, method, requestEntity, String.class);

        return ForwardResult.builder()
                .response(response)
                .requestHeaders(headers)
                .build();
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

    private void handleException(ErrorResult errorResult, CachedBodyHttpServletResponse response) throws IOException {
        HttpStatusCodeException e = errorResult.getException();
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
                                ForwardResult result) {
        Map<String, String> requestHeaders = convertToMap(result.getRequestHeaders());
        Map<String, String> responseHeaders = convertToMap(result.getResponse().getHeaders());

        RecordedExchange exchange = RecordedExchange.builder()
                .method(request.getMethod())
                .path(request.getRequestURI())
                .requestHeaders(requestHeaders)
                .requestBody(request.getCachedBodyAsString())
                .responseStatus(result.getResponse().getStatusCode().value())
                .responseHeaders(responseHeaders)
                .responseBody(result.getResponse().getBody())
                .build();

        recordingService.record(exchange);
    }

    private void recordExchange(CachedBodyHttpServletRequest request,
                                CachedBodyHttpServletResponse response,
                                ErrorResult errorResult) {
        HttpStatusCodeException e = errorResult.getException();
        Map<String, String> requestHeaders = convertToMap(errorResult.getRequestHeaders());
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
