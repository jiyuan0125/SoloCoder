package com.example.messagequeue.service;

import com.example.messagequeue.config.MessageQueueProperties;
import com.example.messagequeue.exception.InvalidJsonException;
import com.example.messagequeue.exception.MessageTooLargeException;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.core.JsonToken;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.nio.charset.StandardCharsets;

@Slf4j
@Service
@RequiredArgsConstructor
public class MessageValidationService {

    private final MessageQueueProperties properties;
    private final ObjectMapper objectMapper;

    public void validateMessage(String content, String contentType) {
        validateSize(content);
        if (isJsonContentType(contentType)) {
            validateJson(content);
        }
    }

    private void validateSize(String content) {
        if (content == null) {
            return;
        }
        int size = content.getBytes(StandardCharsets.UTF_8).length;
        int maxSize = properties.getMaxMessageSizeBytes();
        if (size > maxSize) {
            log.warn("Message size {} exceeds maximum {}", size, maxSize);
            throw new MessageTooLargeException(size, maxSize);
        }
    }

    private boolean isJsonContentType(String contentType) {
        if (contentType == null) {
            return false;
        }
        String lower = contentType.toLowerCase();
        return lower.contains("application/json") || lower.endsWith("+json");
    }

    private void validateJson(String content) {
        if (content == null || content.isEmpty()) {
            return;
        }
        try {
            objectMapper.readTree(content);
        } catch (JsonProcessingException e) {
            int line = e.getLocation() != null ? e.getLocation().getLineNr() : -1;
            int column = e.getLocation() != null ? e.getLocation().getColumnNr() : -1;
            String msg = e.getOriginalMessage() != null ? e.getOriginalMessage() : "Invalid JSON";
            log.warn("Invalid JSON at line {}, column {}: {}", line, column, msg);
            throw new InvalidJsonException(line, column, msg);
        }
    }
}
