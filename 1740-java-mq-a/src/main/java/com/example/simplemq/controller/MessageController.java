package com.example.simplemq.controller;

import com.example.simplemq.model.Message;
import com.example.simplemq.storage.MessageStorage;
import lombok.Data;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.io.IOException;
import java.util.Map;

@Slf4j
@RestController
@RequestMapping("/api/messages")
@RequiredArgsConstructor
public class MessageController {

    private final MessageStorage messageStorage;

    @PostMapping("/{topicName}/produce")
    public ResponseEntity<?> produceMessage(
            @PathVariable String topicName,
            @RequestBody ProduceMessageRequest request
    ) {
        log.debug("Producing message to topic: {}", topicName);
        try {
            Message message = messageStorage.produceMessage(topicName, request.getContent());
            return ResponseEntity.ok(message);
        } catch (IllegalArgumentException e) {
            log.warn("Failed to produce message: {}", e.getMessage());
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        } catch (IOException e) {
            log.error("Error producing message", e);
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    @GetMapping("/{topicName}/consume")
    public ResponseEntity<?> consumeMessage(
            @PathVariable String topicName,
            @RequestParam String groupId
    ) {
        log.debug("Consuming message from topic: {}, group: {}", topicName, groupId);
        try {
            Message message = messageStorage.consumeMessage(topicName, groupId);
            if (message == null) {
                return ResponseEntity.ok(Map.of("message", "No messages available"));
            }
            return ResponseEntity.ok(message);
        } catch (IllegalArgumentException e) {
            log.warn("Failed to consume message: {}", e.getMessage());
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    @PostMapping("/{topicName}/ack")
    public ResponseEntity<?> acknowledgeMessage(
            @PathVariable String topicName,
            @RequestParam String groupId,
            @RequestParam long offset
    ) {
        log.debug("Acknowledging message: topic={}, group={}, offset={}", topicName, groupId, offset);
        try {
            messageStorage.acknowledgeMessage(topicName, groupId, offset);
            return ResponseEntity.ok(Map.of("message", "Acknowledged successfully"));
        } catch (IllegalArgumentException e) {
            log.warn("Failed to acknowledge message: {}", e.getMessage());
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        } catch (IOException e) {
            log.error("Error acknowledging message", e);
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    @Data
    public static class ProduceMessageRequest {
        private String content;
    }
}
