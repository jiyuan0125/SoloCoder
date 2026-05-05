package com.payroll.client.http;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.Map;

public class HttpClient {

    private static final String BASE_URL = "http://localhost:8080";
    private static final int TIMEOUT = 30000;

    private final ObjectMapper objectMapper;

    public HttpClient() {
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
        this.objectMapper.disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
    }

    public String get(String path, Map<String, String> queryParams) throws IOException {
        StringBuilder urlBuilder = new StringBuilder(BASE_URL + path);
        if (queryParams != null && !queryParams.isEmpty()) {
            urlBuilder.append("?");
            boolean first = true;
            for (Map.Entry<String, String> entry : queryParams.entrySet()) {
                if (!first) {
                    urlBuilder.append("&");
                }
                urlBuilder.append(entry.getKey()).append("=").append(encodeParam(entry.getValue()));
                first = false;
            }
        }

        HttpURLConnection connection = null;
        try {
            URL url = new URL(urlBuilder.toString());
            connection = (HttpURLConnection) url.openConnection();
            connection.setRequestMethod("GET");
            connection.setConnectTimeout(TIMEOUT);
            connection.setReadTimeout(TIMEOUT);
            connection.setRequestProperty("Content-Type", "application/json");

            return readResponse(connection);
        } finally {
            if (connection != null) {
                connection.disconnect();
            }
        }
    }

    public String post(String path, Object body) throws IOException {
        HttpURLConnection connection = null;
        try {
            URL url = new URL(BASE_URL + path);
            connection = (HttpURLConnection) url.openConnection();
            connection.setRequestMethod("POST");
            connection.setConnectTimeout(TIMEOUT);
            connection.setReadTimeout(TIMEOUT);
            connection.setRequestProperty("Content-Type", "application/json");
            connection.setDoOutput(true);

            if (body != null) {
                String jsonBody = objectMapper.writeValueAsString(body);
                try (OutputStream os = connection.getOutputStream()) {
                    os.write(jsonBody.getBytes(StandardCharsets.UTF_8));
                }
            }

            return readResponse(connection);
        } finally {
            if (connection != null) {
                connection.disconnect();
            }
        }
    }

    private String readResponse(HttpURLConnection connection) throws IOException {
        int responseCode = connection.getResponseCode();
        InputStream inputStream;
        if (responseCode >= 200 && responseCode < 300) {
            inputStream = connection.getInputStream();
        } else {
            inputStream = connection.getErrorStream();
            if (inputStream == null) {
                return "{\"code\":" + responseCode + ",\"message\":\"HTTP Error: " + responseCode + "\"}";
            }
        }

        try (InputStream is = inputStream) {
            byte[] buffer = new byte[4096];
            int bytesRead;
            StringBuilder result = new StringBuilder();
            while ((bytesRead = is.read(buffer)) != -1) {
                result.append(new String(buffer, 0, bytesRead, StandardCharsets.UTF_8));
            }
            return result.toString();
        }
    }

    private String encodeParam(String value) {
        try {
            return java.net.URLEncoder.encode(value, StandardCharsets.UTF_8);
        } catch (Exception e) {
            return value;
        }
    }

    public ObjectMapper getObjectMapper() {
        return objectMapper;
    }
}
