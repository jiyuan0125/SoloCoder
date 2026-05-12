package com.example.simplemq.controller;

import com.example.simplemq.storage.MessageStorage;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.io.IOException;
import java.util.Map;

@Slf4j
@RestController
@RequestMapping("/api/consumers")
@RequiredArgsConstructor
public class ConsumerController {

    private final MessageStorage messageStorage;

    @GetMapping("/offsets/{topicName}")
    public ResponseEntity<?> getConsumerOffsets(@PathVariable String topicName) {
        log.debug("Getting consumer offsets for topic: {}", topicName);
        try {
            Map<String, Long> offsets = messageStorage.getConsumerOffsets(topicName);
            return ResponseEntity.ok(offsets);
        } catch (IllegalArgumentException e) {
            log.warn("Failed to get consumer offsets: {}", e.getMessage());
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        }
    }

    @PostMapping("/offsets/{topicName}/reset")
    public ResponseEntity<?> resetConsumerOffset(
            @PathVariable String topicName,
            @RequestParam String groupId,
            @RequestParam long offset
    ) {
        log.info("Resetting consumer offset: topic={}, group={}, offset={}", topicName, groupId, offset);
        try {
            messageStorage.resetConsumerOffset(topicName, groupId, offset);
            return ResponseEntity.ok(Map.of(
                    "message", "Consumer offset reset successfully",
                    "topic", topicName,
                    "groupId", groupId,
                    "offset", offset
            ));
        } catch (IllegalArgumentException e) {
            log.warn("Failed to reset consumer offset: {}", e.getMessage());
            return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
        } catch (IOException e) {
            log.error("Error resetting consumer offset", e);
            return ResponseEntity.internalServerError().body(Map.of("error", e.getMessage()));
        }
    }
}
