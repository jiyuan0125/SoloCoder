package com.servicemesh.sidecar.service;

import com.servicemesh.common.model.DiscoveryInstance;
import com.servicemesh.sidecar.config.SidecarProperties;
import okhttp3.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.BufferedReader;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.Enumeration;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;

@Service
public class ReverseProxyService {

    private static final Logger log = LoggerFactory.getLogger(ReverseProxyService.class);

    private final SidecarProperties properties;
    private final DiscoveryService discoveryService;
    private final TelemetryService telemetryService;
    private final OkHttpClient httpClient;

    public ReverseProxyService(SidecarProperties properties,
                               DiscoveryService discoveryService,
                               TelemetryService telemetryService) {
        this.properties = properties;
        this.discoveryService = discoveryService;
        this.telemetryService = telemetryService;
        this.httpClient = new OkHttpClient.Builder()
                .connectTimeout(10, TimeUnit.SECONDS)
                .readTimeout(30, TimeUnit.SECONDS)
                .writeTimeout(30, TimeUnit.SECONDS)
                .build();
    }

    public void proxyInbound(HttpServletRequest request, HttpServletResponse response) throws Exception {
        String targetUrl = "http://" + properties.getAppHost() + ":" + properties.getAppPort() + request.getRequestURI();
        if (request.getQueryString() != null) {
            targetUrl += "?" + request.getQueryString();
        }

        String method = request.getMethod();
        RequestBody body = null;
        if (!"GET".equals(method) && !"HEAD".equals(method)) {
            byte[] bodyBytes = readRequestBody(request);
            String contentType = request.getContentType();
            if (contentType == null) contentType = "application/octet-stream";
            body = RequestBody.create(bodyBytes, MediaType.parse(contentType));
        }

        Request.Builder builder = new Request.Builder().url(targetUrl).method(method, body);
        copyHeaders(request, builder);

        long startTime = System.currentTimeMillis();
        boolean success = true;

        try (Response upstream = httpClient.newCall(builder.build()).execute()) {
            response.setStatus(upstream.code());

            for (Map.Entry<String, List<String>> header : upstream.headers().toMultimap().entrySet()) {
                String name = header.getKey();
                if (!"Transfer-Encoding".equalsIgnoreCase(name) && !"Connection".equalsIgnoreCase(name)) {
                    for (String value : header.getValue()) {
                        response.addHeader(name, value);
                    }
                }
            }

            if (upstream.body() != null) {
                try (InputStream is = upstream.body().byteStream();
                     OutputStream os = response.getOutputStream()) {
                    byte[] buffer = new byte[8192];
                    int read;
                    while ((read = is.read(buffer)) != -1) {
                        os.write(buffer, 0, read);
                    }
                }
            }

            success = upstream.code() < 500;
        } catch (Exception e) {
            success = false;
            response.setStatus(503);
            response.getWriter().write("Service Unavailable: " + e.getMessage());
            log.warn("Inbound proxy error", e);
        }
    }

    public void proxyOutboundWithPath(HttpServletRequest request, HttpServletResponse response, String path) throws Exception {
        String serviceName = extractServiceName(path);

        if (serviceName == null) {
            response.setStatus(400);
            response.getWriter().write("Missing service name in path");
            return;
        }

        discoveryService.discoverService(serviceName);

        DiscoveryInstance instance = discoveryService.selectInstance(serviceName);
        if (instance == null) {
            response.setStatus(503);
            response.getWriter().write("No healthy instance available for: " + serviceName);
            return;
        }

        String remainingPath = path.substring(serviceName.length() + 1);
        if (!remainingPath.startsWith("/")) remainingPath = "/" + remainingPath;

        String targetUrl = "http://" + instance.getIp() + ":" + instance.getPort() + remainingPath;
        if (request.getQueryString() != null) {
            targetUrl += "?" + request.getQueryString();
        }

        String method = request.getMethod();
        RequestBody body = null;
        if (!"GET".equals(method) && !"HEAD".equals(method)) {
            byte[] bodyBytes = readRequestBody(request);
            String contentType = request.getContentType();
            if (contentType == null) contentType = "application/octet-stream";
            body = RequestBody.create(bodyBytes, MediaType.parse(contentType));
        }

        Request.Builder builder = new Request.Builder().url(targetUrl).method(method, body);
        copyHeaders(request, builder);

        long startTime = System.currentTimeMillis();
        boolean success = true;

        try (Response upstream = httpClient.newCall(builder.build()).execute()) {
            response.setStatus(upstream.code());

            for (Map.Entry<String, List<String>> header : upstream.headers().toMultimap().entrySet()) {
                String name = header.getKey();
                if (!"Transfer-Encoding".equalsIgnoreCase(name) && !"Connection".equalsIgnoreCase(name)) {
                    for (String value : header.getValue()) {
                        response.addHeader(name, value);
                    }
                }
            }

            if (upstream.body() != null) {
                try (InputStream is = upstream.body().byteStream();
                     OutputStream os = response.getOutputStream()) {
                    byte[] buffer = new byte[8192];
                    int read;
                    while ((read = is.read(buffer)) != -1) {
                        os.write(buffer, 0, read);
                    }
                }
            }

            success = upstream.code() < 500;
        } catch (Exception e) {
            success = false;
            response.setStatus(503);
            response.getWriter().write("Upstream error: " + e.getMessage());
            log.warn("Outbound proxy error to {}: {}", serviceName, e.getMessage());
        } finally {
            long latency = System.currentTimeMillis() - startTime;
            telemetryService.recordCall(properties.getServiceName(), serviceName, success, latency);
        }
    }

    private String extractServiceName(String path) {
        if (path.startsWith("/")) path = path.substring(1);
        int idx = path.indexOf("/");
        return idx > 0 ? path.substring(0, idx) : (path.isEmpty() ? null : path);
    }

    private byte[] readRequestBody(HttpServletRequest request) throws Exception {
        java.io.ByteArrayOutputStream baos = new java.io.ByteArrayOutputStream();
        try (InputStream is = request.getInputStream()) {
            byte[] buffer = new byte[8192];
            int read;
            while ((read = is.read(buffer)) != -1) {
                baos.write(buffer, 0, read);
            }
        }
        return baos.toByteArray();
    }

    private void copyHeaders(HttpServletRequest request, Request.Builder builder) {
        Enumeration<String> names = request.getHeaderNames();
        while (names.hasMoreElements()) {
            String name = names.nextElement();
            if ("Host".equalsIgnoreCase(name) || "Connection".equalsIgnoreCase(name) ||
                "Keep-Alive".equalsIgnoreCase(name)) {
                continue;
            }
            Enumeration<String> values = request.getHeaders(name);
            while (values.hasMoreElements()) {
                builder.addHeader(name, values.nextElement());
            }
        }
    }
}
