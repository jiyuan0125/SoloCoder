package com.example.proxy.controller;

import com.example.proxy.service.RewriteLogService;
import com.example.proxy.service.RewriteRuleService;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.*;
import org.springframework.util.StreamUtils;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.HttpStatusCodeException;
import org.springframework.web.client.ResourceAccessException;
import org.springframework.web.client.RestTemplate;

import javax.servlet.http.HttpServletRequest;
import java.io.IOException;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.Enumeration;
import java.util.Optional;

@RestController
public class ProxyController {

    private final RestTemplate restTemplate;
    private final RewriteRuleService ruleService;
    private final RewriteLogService logService;

    @Value("${proxy.target-url}")
    private String targetUrl;

    public ProxyController(RestTemplate restTemplate,
                           RewriteRuleService ruleService,
                           RewriteLogService logService) {
        this.restTemplate = restTemplate;
        this.ruleService = ruleService;
        this.logService = logService;
    }

    @RequestMapping("/**")
    public ResponseEntity<byte[]> proxy(HttpServletRequest request) throws IOException, URISyntaxException {
        String originalPath = buildOriginalPath(request);

        Optional<RewriteRuleService.RewriteResult> rewriteResult = ruleService.matchAndRewrite(originalPath);

        String targetPath;
        if (rewriteResult.isPresent()) {
            targetPath = rewriteResult.get().getRewrittenPath();
            logService.log(rewriteResult.get().getRuleId(), originalPath, targetPath);
        } else {
            targetPath = originalPath;
        }

        return forwardRequest(request, targetPath);
    }

    private String buildOriginalPath(HttpServletRequest request) {
        String path = request.getRequestURI();
        String queryString = request.getQueryString();
        if (queryString != null && !queryString.isEmpty()) {
            return path + "?" + queryString;
        }
        return path;
    }

    private ResponseEntity<byte[]> forwardRequest(HttpServletRequest request, String targetPath)
            throws URISyntaxException, IOException {
        String method = request.getMethod();
        HttpMethod httpMethod = HttpMethod.valueOf(method);

        HttpHeaders headers = extractHeaders(request);

        byte[] body = StreamUtils.copyToByteArray(request.getInputStream());

        URI targetUri = buildTargetUri(targetPath);

        HttpEntity<byte[]> httpEntity = new HttpEntity<>(body, headers);

        try {
            ResponseEntity<byte[]> response = restTemplate.exchange(
                    targetUri,
                    httpMethod,
                    httpEntity,
                    byte[].class);

            return buildResponse(response);
        } catch (HttpStatusCodeException e) {
            return buildErrorResponse(e);
        } catch (ResourceAccessException e) {
            return buildGatewayErrorResponse();
        }
    }

    private HttpHeaders extractHeaders(HttpServletRequest request) {
        HttpHeaders headers = new HttpHeaders();
        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String headerName = headerNames.nextElement();
            if (!headerName.equalsIgnoreCase("host")) {
                Enumeration<String> headerValues = request.getHeaders(headerName);
                while (headerValues.hasMoreElements()) {
                    headers.add(headerName, headerValues.nextElement());
                }
            }
        }
        return headers;
    }

    private URI buildTargetUri(String targetPath) throws URISyntaxException {
        String baseUrl = targetUrl;
        if (baseUrl.endsWith("/")) {
            baseUrl = baseUrl.substring(0, baseUrl.length() - 1);
        }
        if (!targetPath.startsWith("/")) {
            targetPath = "/" + targetPath;
        }
        return new URI(baseUrl + targetPath);
    }

    private ResponseEntity<byte[]> buildResponse(ResponseEntity<byte[]> response) {
        HttpHeaders responseHeaders = new HttpHeaders();
        response.getHeaders().forEach((key, values) -> {
            if (!isHopByHopHeader(key)) {
                responseHeaders.addAll(key, values);
            }
        });
        return ResponseEntity.status(response.getStatusCode())
                .headers(responseHeaders)
                .body(response.getBody());
    }

    private ResponseEntity<byte[]> buildErrorResponse(HttpStatusCodeException e) {
        HttpHeaders responseHeaders = new HttpHeaders();
        e.getResponseHeaders().forEach((key, values) -> {
            if (!isHopByHopHeader(key)) {
                responseHeaders.addAll(key, values);
            }
        });
        return ResponseEntity.status(e.getStatusCode())
                .headers(responseHeaders)
                .body(e.getResponseBodyAsByteArray());
    }

    private ResponseEntity<byte[]> buildGatewayErrorResponse() {
        return ResponseEntity.status(HttpStatus.BAD_GATEWAY)
                .contentType(MediaType.TEXT_PLAIN)
                .body("Bad Gateway: Could not reach backend server".getBytes());
    }

    private boolean isHopByHopHeader(String headerName) {
        String lower = headerName.toLowerCase();
        return lower.equals("connection")
                || lower.equals("keep-alive")
                || lower.equals("proxy-authenticate")
                || lower.equals("proxy-authorization")
                || lower.equals("te")
                || lower.equals("trailers")
                || lower.equals("transfer-encoding")
                || lower.equals("upgrade");
    }
}
