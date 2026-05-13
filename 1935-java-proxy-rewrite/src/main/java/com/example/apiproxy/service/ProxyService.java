package com.example.apiproxy.service;

import com.example.apiproxy.config.ProxyProperties;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.client.HttpClientErrorException;
import org.springframework.web.client.HttpServerErrorException;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.client.RestTemplate;

import javax.servlet.http.HttpServletRequest;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.*;

@Service
public class ProxyService {

    private static final Logger logger = LoggerFactory.getLogger(ProxyService.class);
    private static final String X_API_VERSION_HEADER = "X-API-Version";
    private static final String X_FORWARDED_FOR_HEADER = "X-Forwarded-For";

    private final ProxyProperties properties;
    private final LoadBalancerService loadBalancer;
    private final CacheService cacheService;
    private final RestTemplate restTemplate;

    public ProxyService(ProxyProperties properties,
                        LoadBalancerService loadBalancer,
                        CacheService cacheService,
                        RestTemplate restTemplate) {
        this.properties = properties;
        this.loadBalancer = loadBalancer;
        this.cacheService = cacheService;
        this.restTemplate = restTemplate;
    }

    public ResponseEntity<byte[]> forward(HttpServletRequest request) throws IOException, URISyntaxException {
        String method = request.getMethod().toUpperCase();
        String path = request.getRequestURI();
        String queryString = request.getQueryString();
        String version = extractVersion(request);

        logger.info("Processing {} request: {}?{} [version: {}]", method, path, queryString, version);

        if ("GET".equals(method) && properties.getCache().isEnabled()) {
            String cacheKey = cacheService.buildKey(version, path, queryString);
            CacheService.CacheEntry cached = cacheService.get(cacheKey);
            if (cached != null) {
                return buildResponseFromCache(cached);
            }
        }

        byte[] body = readRequestBody(request);
        HttpHeaders headers = copyHeaders(request);
        String clientIp = getClientIp(request);
        addXForwardedFor(headers, clientIp);

        ResponseEntity<byte[]> response = executeWithRetry(version, method, path, queryString, headers, body);

        if ("GET".equals(method) && properties.getCache().isEnabled() && isSuccessResponse(response)) {
            String cacheKey = cacheService.buildKey(version, path, queryString);
            CacheService.CacheEntry entry = new CacheService.CacheEntry(
                    response.getStatusCodeValue(),
                    response.getHeaders(),
                    response.getBody()
            );
            cacheService.put(cacheKey, entry);
        }

        return response;
    }

    private ResponseEntity<byte[]> executeWithRetry(String version, String method, String path,
                                                     String queryString, HttpHeaders headers, byte[] body)
            throws URISyntaxException {
        int maxAttempts = properties.getRetry().getMaxAttempts() + 1;
        String lastBackend = null;
        ResponseEntity<byte[]> lastResponse = null;
        Exception lastException = null;

        for (int attempt = 0; attempt < maxAttempts; attempt++) {
            String backend;
            if (attempt == 0) {
                backend = loadBalancer.selectBackend(version);
            } else {
                backend = loadBalancer.selectBackendExcluding(version, lastBackend);
                if (backend == null) {
                    backend = lastBackend;
                    logger.warn("No alternative backend available for version {}, retrying same backend", version);
                }
                logger.warn("Retry attempt {}/{} for version {}, switching from {} to {}",
                        attempt, maxAttempts - 1, version, lastBackend, backend);
            }

            lastBackend = backend;

            try {
                ResponseEntity<byte[]> response = executeRequest(backend, method, path, queryString, headers, body);

                if (is5xxResponse(response)) {
                    logger.error("Backend {} returned 5xx status {} for {} {}, attempt {}/{}",
                            backend, response.getStatusCodeValue(), method, path, attempt + 1, maxAttempts);
                    lastResponse = response;
                    continue;
                }

                if (attempt > 0) {
                    logger.info("Retry succeeded on backend {} after {} attempt(s)", backend, attempt);
                }
                return response;

            } catch (ResourceAccessException e) {
                logger.error("Backend {} connection failed for {} {}, attempt {}/{}: {}",
                        backend, method, path, attempt + 1, maxAttempts, e.getMessage());
                lastException = e;
            } catch (HttpServerErrorException e) {
                logger.error("Backend {} returned 5xx error for {} {}, attempt {}/{}: {}",
                        backend, method, path, attempt + 1, maxAttempts, e.getMessage());
                lastResponse = ResponseEntity.status(e.getRawStatusCode())
                        .headers(e.getResponseHeaders())
                        .body(e.getResponseBodyAsByteArray());
            } catch (HttpClientErrorException e) {
                return ResponseEntity.status(e.getRawStatusCode())
                        .headers(e.getResponseHeaders())
                        .body(e.getResponseBodyAsByteArray());
            }
        }

        if (lastResponse != null) {
            logger.error("All {} attempts failed for version {}, returning last 5xx response", maxAttempts, version);
            return lastResponse;
        }

        if (lastException != null) {
            logger.error("All {} attempts failed for version {}, returning 503", maxAttempts, version);
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                    .body(("Service Unavailable: " + lastException.getMessage()).getBytes());
        }

        return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE).build();
    }

    private ResponseEntity<byte[]> executeRequest(String backend, String method, String path,
                                                   String queryString, HttpHeaders headers, byte[] body)
            throws URISyntaxException {
        String uriString = backend + path;
        if (queryString != null && !queryString.isEmpty()) {
            uriString += "?" + queryString;
        }
        URI uri = new URI(uriString);

        HttpMethod httpMethod = HttpMethod.resolve(method);
        if (httpMethod == null) {
            throw new IllegalArgumentException("Unsupported HTTP method: " + method);
        }

        HttpEntity<byte[]> entity = new HttpEntity<>(body, headers);

        logger.debug("Forwarding {} to {}", method, uri);
        return restTemplate.exchange(uri, httpMethod, entity, byte[].class);
    }

    private String extractVersion(HttpServletRequest request) {
        String version = request.getHeader(X_API_VERSION_HEADER);
        if (version == null || version.trim().isEmpty()) {
            version = properties.getDefaultVersion();
            logger.debug("No X-API-Version header, using default: {}", version);
        }
        if (!properties.getBackends().containsKey(version)) {
            logger.warn("Unknown version {}, falling back to default: {}", version, properties.getDefaultVersion());
            version = properties.getDefaultVersion();
        }
        return version;
    }

    private HttpHeaders copyHeaders(HttpServletRequest request) {
        HttpHeaders headers = new HttpHeaders();
        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String name = headerNames.nextElement();
            if (isHeaderToForward(name)) {
                Enumeration<String> values = request.getHeaders(name);
                List<String> valueList = Collections.list(values);
                headers.put(name, valueList);
            }
        }
        return headers;
    }

    private boolean isHeaderToForward(String name) {
        String lower = name.toLowerCase();
        return !lower.equals("host") &&
               !lower.equals("content-length") &&
               !lower.equals("connection");
    }

    private void addXForwardedFor(HttpHeaders headers, String clientIp) {
        String existing = headers.getFirst(X_FORWARDED_FOR_HEADER);
        if (existing != null && !existing.isEmpty()) {
            headers.set(X_FORWARDED_FOR_HEADER, existing + ", " + clientIp);
        } else {
            headers.set(X_FORWARDED_FOR_HEADER, clientIp);
        }
    }

    private String getClientIp(HttpServletRequest request) {
        String forwardedFor = request.getHeader("X-Forwarded-For");
        if (forwardedFor != null && !forwardedFor.isEmpty()) {
            return forwardedFor.split(",")[0].trim();
        }
        String realIp = request.getHeader("X-Real-IP");
        if (realIp != null && !realIp.isEmpty()) {
            return realIp;
        }
        return request.getRemoteAddr();
    }

    private byte[] readRequestBody(HttpServletRequest request) throws IOException {
        InputStream is = request.getInputStream();
        ByteArrayOutputStream os = new ByteArrayOutputStream();
        byte[] buffer = new byte[4096];
        int read;
        while ((read = is.read(buffer)) != -1) {
            os.write(buffer, 0, read);
        }
        return os.toByteArray();
    }

    private ResponseEntity<byte[]> buildResponseFromCache(CacheService.CacheEntry entry) {
        HttpHeaders headers = new HttpHeaders();
        if (entry.headers != null) {
            headers.putAll(entry.headers);
        }
        return ResponseEntity.status(entry.statusCode)
                .headers(headers)
                .body(entry.body);
    }

    private boolean is5xxResponse(ResponseEntity<?> response) {
        return response.getStatusCodeValue() >= 500 && response.getStatusCodeValue() < 600;
    }

    private boolean isSuccessResponse(ResponseEntity<?> response) {
        int status = response.getStatusCodeValue();
        return status >= 200 && status < 400;
    }
}
