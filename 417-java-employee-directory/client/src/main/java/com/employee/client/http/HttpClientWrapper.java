package com.employee.client.http;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

public class HttpClientWrapper {
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;
    private final String baseUrl;
    private String currentUserId;
    private String currentUserName;
    private String currentRole;

    public HttpClientWrapper(String baseUrl) {
        this.httpClient = HttpClient.newHttpClient();
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
        this.baseUrl = baseUrl;
        this.currentUserId = "E001";
        this.currentUserName = "默认用户";
        this.currentRole = "EMPLOYEE";
    }

    public void setCurrentUser(String userId, String userName, String role) {
        this.currentUserId = userId;
        this.currentUserName = userName;
        this.currentRole = role;
    }

    public <T> T get(String path, TypeReference<T> typeReference) throws IOException, InterruptedException {
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + path))
                .header("Content-Type", "application/json")
                .header("X-User-Id", currentUserId)
                .header("X-User-Name", currentUserName)
                .header("X-User-Role", currentRole)
                .GET()
                .build();

        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        return objectMapper.readValue(response.body(), typeReference);
    }

    public <T> T post(String path, Object body, TypeReference<T> typeReference) throws IOException, InterruptedException {
        String jsonBody = objectMapper.writeValueAsString(body);
        
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + path))
                .header("Content-Type", "application/json")
                .header("X-User-Id", currentUserId)
                .header("X-User-Name", currentUserName)
                .header("X-User-Role", currentRole)
                .POST(HttpRequest.BodyPublishers.ofString(jsonBody))
                .build();

        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        return objectMapper.readValue(response.body(), typeReference);
    }

    public <T> T put(String path, Object body, TypeReference<T> typeReference) throws IOException, InterruptedException {
        String jsonBody = objectMapper.writeValueAsString(body);
        
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + path))
                .header("Content-Type", "application/json")
                .header("X-User-Id", currentUserId)
                .header("X-User-Name", currentUserName)
                .header("X-User-Role", currentRole)
                .PUT(HttpRequest.BodyPublishers.ofString(jsonBody))
                .build();

        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        return objectMapper.readValue(response.body(), typeReference);
    }
}
