package com.loadbalancer.controller;

import com.loadbalancer.model.Instance;
import com.loadbalancer.service.LoadBalancerService;
import jakarta.servlet.http.HttpServletRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.*;
import org.springframework.util.StreamUtils;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.client.RestClient;
import org.springframework.web.client.RestClientException;
import org.springframework.web.util.UriComponentsBuilder;

import java.io.IOException;
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

    @RequestMapping
    public ResponseEntity<byte[]> proxyRequest(HttpServletRequest request) {
        Instance instance = loadBalancerService.getNextInstance();

        if (instance == null) {
            return ResponseEntity.status(HttpStatus.SERVICE_UNAVAILABLE)
                    .body(Collections.singletonMap("error", "No available backend instances")
                            .toString().getBytes());
        }

        instance.incrementActiveRequests();

        try {
            URI targetUri = buildTargetUri(instance, request);
            HttpHeaders headers = buildHeaders(request);
            byte[] requestBody = readRequestBody(request);

            logger.debug("Proxying request {} {} to {}", request.getMethod(), 
                    request.getRequestURI(), targetUri);

            RestClient.RequestBodySpec requestSpec = restClient.method(HttpMethod.valueOf(request.getMethod()))
                    .uri(targetUri)
                    .headers(httpHeaders -> httpHeaders.addAll(headers));

            if (requestBody != null && requestBody.length > 0) {
                requestSpec.body(requestBody);
            }

            ResponseEntity<byte[]> response = requestSpec.retrieve()
                    .toEntity(byte[].class);

            loadBalancerService.recordRequestSuccess(instance);
            return ResponseEntity.status(response.getStatusCode())
                    .headers(response.getHeaders())
                    .body(response.getBody());

        } catch (RestClientException e) {
            logger.error("Error proxying request to {}: {}", instance.getId(), e.getMessage());
            loadBalancerService.recordRequestFailure(instance);
            return ResponseEntity.status(HttpStatus.BAD_GATEWAY)
                    .body(("{\"error\": \"Backend service error: " + e.getMessage() + "\"}").getBytes());
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
        headers.remove(HttpHeaders.CONTENT_LENGTH);

        return headers;
    }

    private byte[] readRequestBody(HttpServletRequest request) {
        try {
            return StreamUtils.copyToByteArray(request.getInputStream());
        } catch (IOException e) {
            logger.warn("Failed to read request body: {}", e.getMessage());
            return new byte[0];
        }
    }
}
