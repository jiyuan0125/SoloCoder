package com.example.simplemq.controller;

import com.example.simplemq.model.TopicConfig;
import com.example.simplemq.storage.MessageStorage;
import lombok.Data;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.Set;

@Slf4j
@RestController
@RequestMapping("/api/topics")
@RequiredArgsConstructor
public class TopicController {

    private final MessageStorage messageStorage;

    @PostMapping
    public ResponseEntity<?> createTopic(@RequestBody CreateTopicRequest request) {
        log.info("Creating topic: {}", request.getName());
        try {
            TopicConfig config = messageStorage.createTopic(
                    request.getName(),
                    request.getMaxMessages(),
                    request.getRetentionMs()
            );
            return ResponseEntity.ok(config);
        } catch (IllegalArgumentException e) {
            log.warn("Failed to create topic: {}", e.getMessage());
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        } catch (IOException e) {
            log.error("Error creating topic", e);
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    @DeleteMapping("/{topicName}")
    public ResponseEntity<?> deleteTopic(@PathVariable String topicName) {
        log.info("Deleting topic: {}", topicName);
        try {
            messageStorage.deleteTopic(topicName);
            return ResponseEntity.ok(Map.of("message", "Topic deleted successfully"));
        } catch (IllegalArgumentException e) {
            log.warn("Failed to delete topic: {}", e.getMessage());
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        } catch (IOException e) {
            log.error("Error deleting topic", e);
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    @PutMapping("/{topicName}")
    public ResponseEntity<?> updateTopicConfig(
            @PathVariable String topicName,
            @RequestBody UpdateTopicConfigRequest request
    ) {
        log.info("Updating topic config: {}", topicName);
        try {
            TopicConfig config = messageStorage.updateTopicConfig(
                    topicName,
                    request.getMaxMessages(),
                    request.getRetentionMs()
            );
            return ResponseEntity.ok(config);
        } catch (IllegalArgumentException e) {
            log.warn("Failed to update topic config: {}", e.getMessage());
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        } catch (IOException e) {
            log.error("Error updating topic config", e);
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }

    @GetMapping
    public ResponseEntity<Set<String>> listTopics() {
        return ResponseEntity.ok(messageStorage.listTopics());
    }

    @GetMapping("/{topicName}")
    public ResponseEntity<?> getTopicInfo(@PathVariable String topicName) {
        TopicConfig config = messageStorage.getTopicConfig(topicName);
        if (config == null) {
            return ResponseEntity.notFound().build();
        }

        Map<String, Object> info = new HashMap<>();
        info.put("config", config);
        info.put("messageCount", messageStorage.getTopicMessageCount(topicName));
        info.put("consumerOffsets", messageStorage.getConsumerOffsets(topicName));

        return ResponseEntity.ok(info);
    }

    @Data
    public static class CreateTopicRequest {
        private String name;
        private Long maxMessages;
        private Long retentionMs;
    }

    @Data
    public static class UpdateTopicConfigRequest {
        private Long maxMessages;
        private Long retentionMs;
    }
}
