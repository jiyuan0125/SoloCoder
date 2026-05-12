package com.example.gateway.service;

import com.example.gateway.model.BackendTarget;
import com.fasterxml.jackson.databind.ObjectMapper;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.SocketTimeoutException;
import java.net.URL;
import java.util.Enumeration;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

@Service
public class ForwardService {

    @Value("${gateway.forward.timeout:10}")
    private int forwardTimeout;

    private final ObjectMapper objectMapper = new ObjectMapper();

    public void forward(HttpServletRequest request, HttpServletResponse response, BackendTarget backend) throws Exception {
        String backendUrl = buildBackendUrl(request, backend);
        
        try {
            URL url = new URL(backendUrl);
            HttpURLConnection connection = (HttpURLConnection) url.openConnection();
            
            int timeoutMillis = (int) TimeUnit.SECONDS.toMillis(forwardTimeout);
            connection.setConnectTimeout(timeoutMillis);
            connection.setReadTimeout(timeoutMillis);
            connection.setRequestMethod(request.getMethod());
            connection.setDoInput(true);
            
            copyRequestHeaders(request, connection);
            
            if (request.getContentLengthLong() > 0) {
                connection.setDoOutput(true);
                try (OutputStream out = connection.getOutputStream()) {
                    copyRequestBody(request, out);
                }
            }
            
            int responseCode = connection.getResponseCode();
            response.setStatus(responseCode);
            
            Map<String, String> responseHeaders = getResponseHeaders(connection);
            for (Map.Entry<String, String> entry : responseHeaders.entrySet()) {
                response.setHeader(entry.getKey(), entry.getValue());
            }
            
            try (InputStream in = (responseCode >= 400) ? connection.getErrorStream() : connection.getInputStream()) {
                if (in != null) {
                    copyResponseBody(in, response);
                }
            }
            
        } catch (SocketTimeoutException e) {
            sendErrorResponse(response, 504, "GATEWAY_TIMEOUT", "转发超时：后端服务响应超时（" + forwardTimeout + "秒）");
        } catch (Exception e) {
            sendErrorResponse(response, 502, "BAD_GATEWAY", "转发失败：后端地址 [" + backendUrl + "] 错误原因 [" + e.getMessage() + "]");
        }
    }

    private String buildBackendUrl(HttpServletRequest request, BackendTarget backend) {
        String backendBase = backend.getUrl();
        if (backendBase.endsWith("/")) {
            backendBase = backendBase.substring(0, backendBase.length() - 1);
        }
        String path = request.getRequestURI();
        String query = request.getQueryString();
        String fullUrl = backendBase + path;
        if (query != null && !query.isEmpty()) {
            fullUrl += "?" + query;
        }
        return fullUrl;
    }

    private void copyRequestHeaders(HttpServletRequest request, HttpURLConnection connection) {
        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String headerName = headerNames.nextElement();
            if (!isHopByHopHeader(headerName)) {
                connection.setRequestProperty(headerName, request.getHeader(headerName));
            }
        }
    }

    private boolean isHopByHopHeader(String headerName) {
        String lower = headerName.toLowerCase();
        return lower.equals("connection") || 
               lower.equals("keep-alive") || 
               lower.equals("proxy-authenticate") || 
               lower.equals("proxy-authorization") || 
               lower.equals("te") || 
               lower.equals("trailers") || 
               lower.equals("transfer-encoding") || 
               lower.equals("upgrade");
    }

    private void copyRequestBody(HttpServletRequest request, OutputStream out) throws IOException {
        try (InputStream in = request.getInputStream()) {
            byte[] buffer = new byte[4096];
            int bytesRead;
            while ((bytesRead = in.read(buffer)) != -1) {
                out.write(buffer, 0, bytesRead);
            }
        }
    }

    private Map<String, String> getResponseHeaders(HttpURLConnection connection) {
        Map<String, String> headers = new HashMap<>();
        int i = 0;
        while (true) {
            String headerName = connection.getHeaderFieldKey(i);
            String headerValue = connection.getHeaderField(i);
            if (headerName == null && headerValue == null) {
                break;
            }
            if (headerName != null && !isHopByHopHeader(headerName)) {
                headers.put(headerName, headerValue);
            }
            i++;
        }
        return headers;
    }

    private void copyResponseBody(InputStream in, HttpServletResponse response) throws IOException {
        try (OutputStream out = response.getOutputStream()) {
            byte[] buffer = new byte[4096];
            int bytesRead;
            while ((bytesRead = in.read(buffer)) != -1) {
                out.write(buffer, 0, bytesRead);
            }
        }
    }

    private void sendErrorResponse(HttpServletResponse response, int status, String code, String message) throws Exception {
        response.setStatus(status);
        response.setContentType("application/json;charset=UTF-8");
        Map<String, String> error = new HashMap<>();
        error.put("error", code);
        error.put("message", message);
        response.getWriter().write(objectMapper.writeValueAsString(error));
    }
}
