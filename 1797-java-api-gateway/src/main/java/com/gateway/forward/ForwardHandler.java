package com.gateway.forward;

import com.gateway.model.BackendServer;
import com.gateway.model.RequestContext;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.*;
import org.springframework.stereotype.Component;
import org.springframework.util.StreamUtils;
import org.springframework.web.client.*;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.ConnectException;
import java.net.SocketTimeoutException;
import java.util.Enumeration;
import java.util.List;

@Slf4j
@Component
public class ForwardHandler {

    private final RestTemplate restTemplate;
    private final int forwardTimeout;

    public ForwardHandler(@Value("${gateway.forward.timeout:10000}") int forwardTimeout) {
        this.forwardTimeout = forwardTimeout;
        this.restTemplate = new RestTemplate();
        ((org.springframework.http.client.SimpleClientHttpRequestFactory) restTemplate.getRequestFactory())
                .setConnectTimeout(forwardTimeout);
        ((org.springframework.http.client.SimpleClientHttpRequestFactory) restTemplate.getRequestFactory())
                .setReadTimeout(forwardTimeout);
    }

    public void forward(RequestContext context) {
        HttpServletRequest request = context.getRequest();
        HttpServletResponse response = context.getResponse();
        BackendServer backend = context.getSelectedBackend();
        
        if (backend == null) {
            throw new ForwardException("BAD_GATEWAY", 502, "没有可用的后端服务器", null);
        }

        String targetUrl = buildTargetUrl(backend.getUrl(), context);
        
        log.debug("转发请求到: {}", targetUrl);

        try {
            HttpMethod method = HttpMethod.resolve(request.getMethod());
            if (method == null) {
                method = HttpMethod.GET;
            }

            HttpHeaders headers = copyHeaders(request);

            byte[] body = readRequestBody(request);

            HttpEntity<byte[]> entity = new HttpEntity<>(body, headers);

            ResponseEntity<byte[]> backendResponse = restTemplate.exchange(
                    targetUrl, method, entity, byte[].class);

            copyResponse(backendResponse, response);

        } catch (ResourceAccessException e) {
            handleForwardError(e, targetUrl);
        } catch (Exception e) {
            log.error("转发请求异常", e);
            throw new ForwardException("BAD_GATEWAY", 502, "转发请求失败: " + e.getMessage(), targetUrl, e);
        }
    }

    private void handleForwardError(ResourceAccessException e, String targetUrl) {
        Throwable cause = e.getCause();
        
        if (cause instanceof SocketTimeoutException) {
            throw new ForwardException("GATEWAY_TIMEOUT", 504,
                    "后端服务响应超时 (" + forwardTimeout + "ms)", targetUrl, e);
        }
        
        if (cause instanceof ConnectException) {
            throw new ForwardException("BAD_GATEWAY", 502,
                    "后端服务不可达: " + cause.getMessage(), targetUrl, e);
        }

        throw new ForwardException("BAD_GATEWAY", 502,
                "后端服务连接失败: " + e.getMessage(), targetUrl, e);
    }

    private String buildTargetUrl(String backendUrl, RequestContext context) {
        String requestPath = context.getRequestPath();
        
        if (backendUrl.endsWith("/")) {
            backendUrl = backendUrl.substring(0, backendUrl.length() - 1);
        }
        
        if (!requestPath.startsWith("/")) {
            requestPath = "/" + requestPath;
        }
        
        return backendUrl + requestPath;
    }

    private HttpHeaders copyHeaders(HttpServletRequest request) {
        HttpHeaders headers = new HttpHeaders();
        
        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String name = headerNames.nextElement();
            
            if ("host".equalsIgnoreCase(name) ||
                "connection".equalsIgnoreCase(name) ||
                "content-length".equalsIgnoreCase(name)) {
                continue;
            }
            
            Enumeration<String> values = request.getHeaders(name);
            while (values.hasMoreElements()) {
                headers.add(name, values.nextElement());
            }
        }
        
        return headers;
    }

    private byte[] readRequestBody(HttpServletRequest request) throws IOException {
        try (InputStream is = request.getInputStream()) {
            return StreamUtils.copyToByteArray(is);
        }
    }

    private void copyResponse(ResponseEntity<byte[]> backendResponse, HttpServletResponse response) throws IOException {
        response.setStatus(backendResponse.getStatusCodeValue());
        
        HttpHeaders responseHeaders = backendResponse.getHeaders();
        for (String name : responseHeaders.keySet()) {
            List<String> values = responseHeaders.get(name);
            if (values != null) {
                for (String value : values) {
                    response.addHeader(name, value);
                }
            }
        }
        
        byte[] body = backendResponse.getBody();
        if (body != null && body.length > 0) {
            try (OutputStream os = response.getOutputStream()) {
                os.write(body);
            }
        }
    }
}
