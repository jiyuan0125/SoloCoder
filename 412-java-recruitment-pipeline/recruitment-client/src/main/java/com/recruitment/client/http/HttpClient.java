package com.recruitment.client.http;

import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.apache.hc.client5.http.classic.methods.HttpGet;
import org.apache.hc.client5.http.classic.methods.HttpPost;
import org.apache.hc.client5.http.impl.classic.CloseableHttpClient;
import org.apache.hc.client5.http.impl.classic.HttpClients;
import org.apache.hc.core5.http.ClassicHttpResponse;
import org.apache.hc.core5.http.HttpEntity;
import org.apache.hc.core5.http.io.entity.EntityUtils;
import org.apache.hc.core5.http.io.entity.StringEntity;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

public class HttpClient {
    private static final String BASE_URL = "http://localhost:8080";
    private final CloseableHttpClient httpClient;
    private final ObjectMapper objectMapper;

    public HttpClient() {
        this.httpClient = HttpClients.createDefault();
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
        this.objectMapper.disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
        this.objectMapper.configure(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES, false);
    }

    public <T> T get(String path, Class<T> responseType) throws IOException {
        HttpGet httpGet = new HttpGet(BASE_URL + path);
        httpGet.setHeader("Content-Type", "application/json");
        
        return httpClient.execute(httpGet, response -> {
            int statusCode = response.getCode();
            HttpEntity entity = response.getEntity();
            String responseBody = entity != null ? EntityUtils.toString(entity, StandardCharsets.UTF_8) : "";
            return objectMapper.readValue(responseBody, responseType);
        });
    }

    public <T> T post(String path, Object requestBody, Class<T> responseType) throws IOException {
        HttpPost httpPost = new HttpPost(BASE_URL + path);
        httpPost.setHeader("Content-Type", "application/json");
        
        if (requestBody != null) {
            String jsonBody = objectMapper.writeValueAsString(requestBody);
            httpPost.setEntity(new StringEntity(jsonBody, StandardCharsets.UTF_8));
        }
        
        return httpClient.execute(httpPost, response -> {
            HttpEntity entity = response.getEntity();
            String responseBody = entity != null ? EntityUtils.toString(entity, StandardCharsets.UTF_8) : "";
            return objectMapper.readValue(responseBody, responseType);
        });
    }

    public void close() throws IOException {
        httpClient.close();
    }
}
