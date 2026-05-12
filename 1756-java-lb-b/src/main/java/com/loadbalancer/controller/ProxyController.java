package com.loadbalancer.controller;

import com.loadbalancer.model.Instance;
import com.loadbalancer.service.LoadBalancerService;
import jakarta.servlet.http.HttpServletRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.*;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.client.RestClient;
import org.springframework.web.client.RestClientException;
import org.springframework.web.util.UriComponentsBuilder;

import java.net.URI;
import java.util.Collections;
import java.util.Enumeration;

@RestController
@RequestMapping("/**")
public class ProxyController {
    private static final Logger logger = LoggerFactory.getLogger(ProxyController.class);

    private final LoadBalancerService loadBalancerService;
    private final RestClient restClient;

    public ProxyController(LoadBalancerService loadBalancerService) {
        this.loadBalancerService = loadBalancerService;
        this.restClient = RestClient.create();
    }

    @GetMapping
    public ResponseEntity<Object> proxyGet(HttpServletRequest request) {
        return proxyRequest(request, HttpMethod.GET, null);
    }

    @PostMapping
    public ResponseEntity<Object> proxyPost(HttpServletRequest request, @RequestBody(required = false) Object body) {
        return proxyRequest(request, HttpMethod.POST, body);
    }

    @PutMapping
    public ResponseEntity<Object> proxyPut(HttpServletRequest request, @RequestBody(required = false) Object body) {
        return proxyRequest(request, HttpMethod.PUT, body);
    }

    @DeleteMapping
    public ResponseEntity<Object> proxyDelete(HttpServletRequest request) {
        return proxyRequest(request, HttpMethod.DELETE, null);
    }

    @PatchMapping
    public ResponseEntity<Object> proxyPatch(HttpServletRequest request, @RequestBody(required = false) Object body) {
        return proxyRequest(request, HttpMethod.PATCH, body);
    }

    private ResponseEntity<Object> proxyRequest(HttpServletRequest request, HttpMethod method, Object body) {
        Instance instance = loadBalancerService.getNextInstance();

        if (instance == null) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                    .body(Collections.singletonMap("error", "No available backend instances"));
        }

        instance.incrementActiveRequests();

        try {
            URI targetUri = buildTargetUri(instance, request);
            HttpHeaders headers = buildHeaders(request);

            logger.debug("Proxying request {} {} to {}", method, request.getRequestURI(), targetUri);

            RestClient.RequestBodySpec requestSpec = restClient.method(method)
                    .uri(targetUri)
                    .headers(httpHeaders -> httpHeaders.addAll(headers));

            if (body != null) {
                requestSpec.body(body);
            }

            ResponseEntity<Object> response = requestSpec.retrieve()
                    .toEntity(Object.class);

            loadBalancerService.recordRequestSuccess(instance);
            return response;

        } catch (RestClientException e) {
            logger.error("Error proxying request to {}: {}", instance.getId(), e.getMessage());
            loadBalancerService.recordRequestFailure(instance);
            return ResponseEntity.status(HttpStatus.BAD_GATEWAY)
                    .body(Collections.singletonMap("error", "Backend service error: " + e.getMessage()));
        } finally {
            instance.decrementActiveRequests();
        }
    }

    private URI buildTargetUri(Instance instance, HttpServletRequest request) {
        String queryString = request.getQueryString();
        String path = request.getRequestURI();

        UriComponentsBuilder builder = UriComponentsBuilder.fromHttpUrl(instance.getUrl()).path(path);

        if (queryString != null && !queryString.isEmpty()) {
            builder.query(queryString);
        }

        return builder.build().toUri();
    }

    private HttpHeaders buildHeaders(HttpServletRequest request) {
        HttpHeaders headers = new HttpHeaders();

        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String headerName = headerNames.nextElement();
            Enumeration<String> values = request.getHeaders(headerName);
            while (values.hasMoreElements()) {
                headers.add(headerName, values.nextElement());
            }
        }

        headers.remove(HttpHeaders.HOST);

        return headers;
    }
}
