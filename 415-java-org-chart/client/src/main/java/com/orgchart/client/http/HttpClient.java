package com.orgchart.client.http;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.apache.commons.httpclient.HttpMethod;
import org.apache.commons.httpclient.methods.DeleteMethod;
import org.apache.commons.httpclient.methods.GetMethod;
import org.apache.commons.httpclient.methods.PostMethod;
import org.apache.commons.httpclient.methods.PutMethod;
import org.apache.commons.httpclient.methods.StringRequestEntity;

import java.io.IOException;
import java.lang.reflect.Type;

public class HttpClient {

    private static final String DEFAULT_BASE_URL = "http://localhost:8080";
    private final String baseUrl;
    private final org.apache.commons.httpclient.HttpClient client;
    private final ObjectMapper objectMapper;

    public HttpClient() {
        this(DEFAULT_BASE_URL);
    }

    public HttpClient(String baseUrl) {
        this.baseUrl = baseUrl;
        this.client = new org.apache.commons.httpclient.HttpClient();
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
        this.objectMapper.disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
    }

    public <T> T get(String path, Type responseType) throws IOException {
        GetMethod method = new GetMethod(baseUrl + path);
        return executeMethod(method, responseType);
    }

    public <T> T get(String path, TypeReference<T> typeRef) throws IOException {
        GetMethod method = new GetMethod(baseUrl + path);
        return executeMethod(method, typeRef);
    }

    public <T> T get(String path, Class<T> responseType) throws IOException {
        GetMethod method = new GetMethod(baseUrl + path);
        return executeMethod(method, responseType);
    }

    public <T> T post(String path, Object requestBody, Type responseType) throws IOException {
        PostMethod method = new PostMethod(baseUrl + path);
        if (requestBody != null) {
            String json = objectMapper.writeValueAsString(requestBody);
            method.setRequestEntity(new StringRequestEntity(json, "application/json", "UTF-8"));
        }
        return executeMethod(method, responseType);
    }

    public <T> T post(String path, Object requestBody, Class<T> responseType) throws IOException {
        PostMethod method = new PostMethod(baseUrl + path);
        if (requestBody != null) {
            String json = objectMapper.writeValueAsString(requestBody);
            method.setRequestEntity(new StringRequestEntity(json, "application/json", "UTF-8"));
        }
        return executeMethod(method, responseType);
    }

    public <T> T put(String path, Object requestBody, Type responseType) throws IOException {
        PutMethod method = new PutMethod(baseUrl + path);
        if (requestBody != null) {
            String json = objectMapper.writeValueAsString(requestBody);
            method.setRequestEntity(new StringRequestEntity(json, "application/json", "UTF-8"));
        }
        return executeMethod(method, responseType);
    }

    public <T> T put(String path, Object requestBody, Class<T> responseType) throws IOException {
        PutMethod method = new PutMethod(baseUrl + path);
        if (requestBody != null) {
            String json = objectMapper.writeValueAsString(requestBody);
            method.setRequestEntity(new StringRequestEntity(json, "application/json", "UTF-8"));
        }
        return executeMethod(method, responseType);
    }

    public <T> T delete(String path, Type responseType) throws IOException {
        DeleteMethod method = new DeleteMethod(baseUrl + path);
        return executeMethod(method, responseType);
    }

    public <T> T delete(String path, Class<T> responseType) throws IOException {
        DeleteMethod method = new DeleteMethod(baseUrl + path);
        return executeMethod(method, responseType);
    }

    private <T> T executeMethod(HttpMethod method, Type responseType) throws IOException {
        try {
            client.executeMethod(method);
            String responseBody = method.getResponseBodyAsString();
            if (responseType == null || responseType == Void.class) {
                return null;
            }
            return objectMapper.readValue(responseBody, objectMapper.constructType(responseType));
        } finally {
            method.releaseConnection();
        }
    }

    private <T> T executeMethod(HttpMethod method, TypeReference<T> typeRef) throws IOException {
        try {
            client.executeMethod(method);
            String responseBody = method.getResponseBodyAsString();
            if (typeRef == null) {
                return null;
            }
            return objectMapper.readValue(responseBody, typeRef);
        } finally {
            method.releaseConnection();
        }
    }

    private <T> T executeMethod(HttpMethod method, Class<T> responseType) throws IOException {
        try {
            client.executeMethod(method);
            String responseBody = method.getResponseBodyAsString();
            if (responseType == null || responseType == Void.class) {
                return null;
            }
            return objectMapper.readValue(responseBody, responseType);
        } finally {
            method.releaseConnection();
        }
    }

    public ObjectMapper getObjectMapper() {
        return objectMapper;
    }
}
